package service

import (
	"context"
	"errors"
	"time"

	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementService struct {
	Repo  *repository.AchievementRepository
	Mongo *mongo.Client
}

func NewAchievementService(repo *repository.AchievementRepository, mongo *mongo.Client) *AchievementService {
	return &AchievementService{
		Repo:  repo,
		Mongo: mongo,
	}
}

// DTO INPUT
type AchievementInput struct {
	AchievementType string                 `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
	Points          float64                `json:"points"`
}

// DTO OUTPUT
type AchievementOutput struct {
	MongoID   string                 `json:"mongo_id"`
	StudentID string                 `json:"student_id"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
}

//              HANDLER (FIBER)
// CREATE
func (s *AchievementService) Create(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)

	var input AchievementInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	result, err := s.CreateAchievement(c.Context(), claims.UserID, claims.Role, input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// SUBMIT (draft → submitted)
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	achievementID := c.Params("id")

	err := s.SubmitAchievement(c.Context(), claims.UserID, claims.Role, achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":        "achievement submitted",
		"achievement_id": achievementID,
		"status":         "submitted",
	})
}

//         INTERNAL BUSINESS LOGIC
// Create Logic
func (s *AchievementService) CreateAchievement(
	ctx context.Context,
	userID int64,
	role string,
	input AchievementInput,
) (*AchievementOutput, error) {

	if role != "mahasiswa" {
		return nil, errors.New("only students can create achievements")
	}

	studentID, err := s.Repo.GetStudentID(ctx, userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Build Mongo document
	doc := map[string]interface{}{
		"studentId":       studentID,
		"achievementType": input.AchievementType,
		"title":           input.Title,
		"description":     input.Description,
		"details":         input.Details,
		"tags":            input.Tags,
		"points":          input.Points,
		"createdAt":       time.Now(),
		"updatedAt":       time.Now(),
	}
	collection := s.Mongo.Database("uas").Collection("achievements")
	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, errors.New("failed to save achievement to mongo")
	}
	objectID := result.InsertedID.(primitive.ObjectID).Hex()
	err = s.Repo.InsertReference(ctx, studentID, objectID)
	if err != nil {
		return nil, errors.New("failed to save reference to postgres")
	}
	return &AchievementOutput{
		MongoID:   objectID,
		StudentID: studentID,
		Status:    "draft",
		Data:      doc,
	}, nil
}

// SUBMIT LOGIC
func (s *AchievementService) SubmitAchievement(
	ctx context.Context,
	userID int64,
	role string,
	achievementID string,
) error {
	if role != "mahasiswa" {
		return errors.New("only students can submit achievements")
	}
	studentID, err := s.Repo.GetStudentID(ctx, userID)
	if err != nil {
		return errors.New("student profile not found")
	}
	err = s.Repo.Submit(ctx, achievementID, studentID)
	if err != nil {
		return err
	}
	return nil
}

func (s *AchievementService) GetMyAchievements(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	ctx := c.Context()

	// 1. Hanya mahasiswa yang boleh melihat prestasinya
	if claims.Role != "mahasiswa" {
		return c.Status(403).JSON(fiber.Map{
			"error": "only students can view their achievements",
		})
	}

	// 2. Ambil student_id dari PostgreSQL
	studentID, err := s.Repo.GetStudentID(ctx, claims.UserID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "student profile not found",
		})
	}

	// 3. Ambil reference dari PostgreSQL
	refs, err := s.Repo.GetByStudentID(ctx, studentID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 4. Gabungkan dengan MongoDB
	collection := s.Mongo.Database("uas").Collection("achievements")

	var results []map[string]interface{}

	for _, ref := range refs {
		mongoID := ref["mongo_id"].(string)

		var mongoDoc map[string]interface{}
		objID, _ := primitive.ObjectIDFromHex(mongoID)

		err := collection.FindOne(ctx, primitive.M{"_id": objID}).Decode(&mongoDoc)
		if err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"reference": ref,
			"mongo":     mongoDoc,
		})
	}
	return c.JSON(results)
}

func (s *AchievementService) GetSupervisedAchievements(c *fiber.Ctx) error {

    claims := c.Locals("claims").(*utils.Claims)
    ctx := c.Context()

    // 1. Hanya dosen wali yang boleh melihat ini
    if claims.Role != "dosen wali" {
        return c.Status(403).JSON(fiber.Map{
            "error": "only academic advisors can view supervised achievements",
        })
    }

    // 2. Ambil lecturer_id berdasarkan user_id
    advisorID, err := s.Repo.GetLecturerID(ctx, claims.UserID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "lecturer profile not found"})
    }

    // 3. Ambil semua mahasiswa bimbingan
    studentIDs, err := s.Repo.GetStudentsByAdvisor(ctx, advisorID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // Jika dosen belum punya mahasiswa bimbingan
    if len(studentIDs) == 0 {
        return c.JSON([]interface{}{})
    }

    // 4. Ambil semua references
    refs, err := s.Repo.GetReferencesByStudentList(ctx, studentIDs)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    // 5. Ambil dokumen mongo untuk masing-masing prestasi
    collection := s.Mongo.Database("uas").Collection("achievements")

    var results []map[string]interface{}
    for _, ref := range refs {
        mongoID := ref["mongo_id"].(string)
        var mongoDoc map[string]interface{}
        objID, _ := primitive.ObjectIDFromHex(mongoID)
        err := collection.FindOne(ctx, primitive.M{"_id": objID}).Decode(&mongoDoc)
        if err != nil {
            continue
        }
        results = append(results, map[string]interface{}{
            "reference": ref,
            "mongo":     mongoDoc,
        })
    }
    return c.JSON(results)
}

func (s *AchievementService) Verify(c *fiber.Ctx) error {
    claims := c.Locals("claims").(*utils.Claims)

    // 1. Hanya dosen wali
    if claims.Role != "dosen wali" {
        return c.Status(403).JSON(fiber.Map{
            "error": "only advisors can verify achievements",
        })
    }

    achievementID := c.Params("id")
    ctx := c.Context()

    // 2. Ambil lecturer_id
    lecturerID, err := s.Repo.GetLecturerID(ctx, claims.UserID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "lecturer profile not found"})
    }

    // 3. Ambil student_id pemilik achievement
    studentID, err := s.Repo.GetStudentIDByAchievement(ctx, achievementID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "achievement not found"})
    }

    // 4. Cek apakah mahasiswa ini adalah bimbingan dosen
    ok, err := s.Repo.IsStudentSupervised(ctx, lecturerID, studentID)
    if err != nil || !ok {
        return c.Status(403).JSON(fiber.Map{
            "error": "student not supervised by this lecturer",
        })
    }

    // 5. Ambil dokumen Mongo untuk melihat poin
    collection := s.Mongo.Database("uas").Collection("achievements")

    var mongoDoc map[string]interface{}
    objID, _ := primitive.ObjectIDFromHex(achievementID)

    err = collection.FindOne(ctx, primitive.M{"_id": objID}).Decode(&mongoDoc)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "mongo document not found"})
    }

    points, _ := mongoDoc["points"].(float64)

    // 6. Update postgres: verified + tambah poin
    err = s.Repo.Verify(ctx, achievementID, studentID, lecturerID, points)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "achievement_id": achievementID,
        "status":         "verified",
        "added_points":   points,
        "message":        "achievement verified successfully",
    })
}
