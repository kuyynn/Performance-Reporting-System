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

// ---------------------------
// DTO INPUT
// ---------------------------
type AchievementInput struct {
	AchievementType string                 `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
	Points          float64                `json:"points"`
}

// ---------------------------
// DTO OUTPUT
// ---------------------------
type AchievementOutput struct {
	MongoID   string                 `json:"mongo_id"`
	StudentID string                 `json:"student_id"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
}

//
//
// =========================================
//              HANDLER (FIBER)
// =========================================
//

// =======================
// CREATE
// =======================
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

// =======================
// SUBMIT (draft → submitted)
// =======================
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

//
//
// =========================================
//         INTERNAL BUSINESS LOGIC
// =========================================
//

// ===============
// Create Logic
// ===============
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

// ===============
// SUBMIT LOGIC
// ===============
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
