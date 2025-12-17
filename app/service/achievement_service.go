package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"os"

	"path/filepath"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
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

//	HANDLER (FIBER)
//
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

func (s *AchievementService) CreateHandler(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	userID := claims.UserID
	role := claims.Role
	var input AchievementInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}
	result, err := s.CreateAchievement(c.Context(), userID, role, input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
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

//	INTERNAL BUSINESS LOGIC
//
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
	refs, err := s.Repo.GetByStudentID(ctx, studentID, false)
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

type RejectInput struct {
	Note string `json:"note"`
}

func (s *AchievementService) Reject(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)

	// 1. Harus dosen wali
	if claims.Role != "dosen wali" {
		return c.Status(403).JSON(fiber.Map{
			"error": "only advisors can reject achievements",
		})
	}

	achievementID := c.Params("id")
	ctx := c.Context()

	// 2. Body Parser
	var input RejectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	if input.Note == "" {
		return c.Status(400).JSON(fiber.Map{"error": "rejection note is required"})
	}

	// 3. Ambil lecturer_id
	lecturerID, err := s.Repo.GetLecturerID(ctx, claims.UserID)
	fmt.Println("DEBUG: claims.UserID =", claims.UserID, " -> lecturerID =", lecturerID, " err =", err)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "lecturer profile not found"})
	}

	// 4. Ambil student_id pemilik achievement
	studentID, err := s.Repo.GetStudentIDByAchievement(ctx, achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not found"})
	}

	// 5. Cek apakah mahasiswa bimbingannya
	ok, err := s.Repo.IsStudentSupervised(ctx, lecturerID, studentID)
	if err != nil || !ok {
		return c.Status(403).JSON(fiber.Map{
			"error": "student not supervised by this lecturer",
		})
	}

	// 6. Jalankan reject (Postgres)
	err = s.Repo.Reject(ctx, achievementID, studentID, lecturerID, input.Note)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"achievement_id": achievementID,
		"status":         "rejected",
		"note":           input.Note,
		"message":        "achievement rejected",
	})
}

func (s *AchievementService) Delete(ctx context.Context, userID int64, role string, achievementID string) error {

	if role != "mahasiswa" {
		return errors.New("only mahasiswa can delete")
	}

	// 1. Soft delete MongoDB (tambahkan deletedAt)
	collection := s.Mongo.Database("uas").Collection("achievements")

	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return errors.New("invalid_mongo_id")
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"deletedAt": time.Now()}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed_soft_delete_mongo")
	}

	// 2. Soft delete reference PostgreSQL
	err = s.Repo.SoftDelete(ctx, achievementID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AchievementService) DeleteHandler(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	userID := claims.UserID
	role := claims.Role
	achievementID := c.Params("id")

	err := s.Delete(c.Context(), userID, role, achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "achievement deleted",
		"id":      achievementID,
	})
}

// ADMIN: VIEW ALL ACHIEVEMENTS WITH FILTERING & PAGINATION
func (s *AchievementService) AdminListAchievements(c *fiber.Ctx) error {
	ctx := c.Context()

	// --- Ambil query params ---
	status := c.Query("status")
	studentID := c.Query("student_id")
	sort := c.Query("sort")
	if sort == "" {
		sort = "created_at"
	}
	order := c.Query("order")
	if order == "" {
		order = "desc"
	}

	// Pagination
	page := 1
	limit := 10
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	offset := (page - 1) * limit

	// --- Panggil repository ---
	filter := repository.AchievementAdminFilter{
		Status:    status,
		StudentID: studentID,
		Sort:      sort,
		Order:     strings.ToLower(order),
		Limit:     limit,
		Offset:    offset,
	}
	refs, total, err := s.Repo.AdminGetAll(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// --- Join dengan MongoDB ---
	collection := s.Mongo.Database("uas").Collection("achievements")
	var results []map[string]interface{}
	for _, ref := range refs {
		mongoID, ok := ref["mongo_id"].(string)
		if !ok || mongoID == "" {
			continue
		}
		objID, err := primitive.ObjectIDFromHex(mongoID)
		if err != nil {
			continue
		}
		var mongoDoc map[string]interface{}
		if err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&mongoDoc); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"reference": ref,
			"mongo":     mongoDoc,
		})
	}

	// --- Pagination meta ---
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return c.JSON(fiber.Map{
		"data": results,
		"meta": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	ctx := c.Context()
	achievementID := c.Params("id")
	// 1. Validasi ObjectId Mongo
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_achievement_id",
		})
	}
	// 2. Ambil reference dari PostgreSQL
	ref, err := s.Repo.GetReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement_not_found",
		})
	}
	studentUUID, _ := ref["student_uuid"].(string)
	// 3. RBAC CHECK
	switch claims.Role {
	case "admin":
	case "mahasiswa":
		myStudentID, err := s.Repo.GetStudentID(ctx, claims.UserID)
		if err != nil || myStudentID != studentUUID {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
	case "dosen wali":
		lecturerID, err := s.Repo.GetLecturerID(ctx, claims.UserID)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "lecturer_profile_not_found",
			})
		}
		ok, err := s.Repo.IsStudentSupervised(ctx, lecturerID, studentUUID)
		if err != nil || !ok {
			return c.Status(403).JSON(fiber.Map{
				"error": "student_not_supervised",
			})
		}
	default:
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
	// 4. Ambil dokumen Mongo
	collection := s.Mongo.Database("uas").Collection("achievements")
	var mongoDoc map[string]interface{}
	if err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&mongoDoc); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "mongo_document_not_found",
		})
	}
	// 5. Response
	return c.JSON(fiber.Map{
		"reference": ref,
		"mongo":     mongoDoc,
	})
}

// UPDATE DRAFT ACHIEVEMENT
func (s *AchievementService) UpdateDraft(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	ctx := c.Context()
	achievementID := c.Params("id")

	// 1. Hanya mahasiswa
	if claims.Role != "mahasiswa" {
		return c.Status(403).JSON(fiber.Map{
			"error": "only students can update achievements",
		})
	}

	// 2. Validasi ObjectId
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_achievement_id",
		})
	}

	// 3. Ambil student_id
	studentID, err := s.Repo.GetStudentID(ctx, claims.UserID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "student_profile_not_found",
		})
	}

	// 4. Cek reference & status
	ref, err := s.Repo.GetReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement_not_found",
		})
	}
	if ref["status"] != "draft" {
		return c.Status(422).JSON(fiber.Map{
			"error": "only_draft_can_be_updated",
		})
	}
	if ref["student_uuid"] != studentID {
		return c.Status(403).JSON(fiber.Map{
			"error": "not_achievement_owner",
		})
	}

	// 5. Body parser (reuse DTO Create)
	var input AchievementInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_request",
		})
	}

	// 6. Update Mongo document
	collection := s.Mongo.Database("uas").Collection("achievements")
	update := bson.M{
		"$set": bson.M{
			"achievementType": input.AchievementType,
			"title":           input.Title,
			"description":     input.Description,
			"details":         input.Details,
			"tags":            input.Tags,
			"points":          input.Points,
			"updatedAt":       time.Now(),
		},
	}
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		update,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_update_mongo",
		})
	}
	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "mongo_document_not_found",
		})
	}

	// 7. Ambil data terbaru untuk response
	var mongoDoc map[string]interface{}
	_ = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&mongoDoc)
	return c.JSON(fiber.Map{
		"message": "achievement updated",
		"id":      achievementID,
		"status":  "draft",
		"mongo":   mongoDoc,
	})
}

// GET ACHIEVEMENT HISTORY
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	ctx := c.Context()
	achievementID := c.Params("id")

	// 1. Validasi ObjectId
	if _, err := primitive.ObjectIDFromHex(achievementID); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_achievement_id",
		})
	}

	// 2. Ambil reference (untuk RBAC)
	ref, err := s.Repo.GetReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement_not_found",
		})
	}
	studentUUID, _ := ref["student_uuid"].(string)

	// 3. RBAC
	switch claims.Role {
	case "admin":
	case "mahasiswa":
		myStudentID, err := s.Repo.GetStudentID(ctx, claims.UserID)
		if err != nil || myStudentID != studentUUID {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
	case "dosen wali":
		lecturerID, err := s.Repo.GetLecturerID(ctx, claims.UserID)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer_profile_not_found"})
		}
		ok, err := s.Repo.IsStudentSupervised(ctx, lecturerID, studentUUID)
		if err != nil || !ok {
			return c.Status(403).JSON(fiber.Map{"error": "student_not_supervised"})
		}
	default:
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	// 4. Ambil history
	history, err := s.Repo.GetHistoryByMongoID(ctx, achievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_history",
		})
	}
	return c.JSON(fiber.Map{
		"achievement_id": achievementID,
		"history":        history,
	})
}

// UPLOAD ATTACHMENT
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	ctx := c.Context()
	achievementID := c.Params("id")

	// 1. Hanya mahasiswa
	if claims.Role != "mahasiswa" {
		return c.Status(403).JSON(fiber.Map{
			"error": "only students can upload attachments",
		})
	}

	// 2. Validasi ObjectId
	objID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid_achievement_id",
		})
	}

	// 3. Ambil student_id
	studentID, err := s.Repo.GetStudentID(ctx, claims.UserID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "student_profile_not_found",
		})
	}

	// 4. Ambil reference & cek ownership
	ref, err := s.Repo.GetReferenceByMongoID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement_not_found",
		})
	}
	if ref["student_uuid"] != studentID {
		return c.Status(403).JSON(fiber.Map{
			"error": "not_achievement_owner",
		})
	}

	// 5. Status check
	if ref["status"] != "draft" && ref["status"] != "submitted" {
		return c.Status(422).JSON(fiber.Map{
			"error": "attachments_not_allowed_for_this_status",
		})
	}

	// 6. Ambil file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "file_required",
		})
	}

	// 7. Validasi tipe file (basic)
	allowed := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowed[ext] {
		return c.Status(422).JSON(fiber.Map{
			"error": "invalid_file_type",
		})
	}

	// 8. Simpan file
	basePath := fmt.Sprintf("./uploads/achievements/%s", achievementID)
	_ = os.MkdirAll(basePath, os.ModePerm)
	savePath := fmt.Sprintf("%s/%s", basePath, file.Filename)
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_save_file",
		})
	}

	// 9. Update MongoDB
	collection := s.Mongo.Database("uas").Collection("achievements")
	attachment := bson.M{
		"filename":   file.Filename,
		"path":       savePath,
		"uploadedAt": time.Now(),
	}
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$push": bson.M{"attachments": attachment}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_update_mongo",
		})
	}
	return c.JSON(fiber.Map{
		"message": "attachment uploaded",
		"file": fiber.Map{
			"name": file.Filename,
			"path": savePath,
		},
	})
}
