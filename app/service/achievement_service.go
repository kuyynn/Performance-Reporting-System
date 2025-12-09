package service

import (
	"context"
	"errors"
	"time"
	
	"uas/app/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementService struct {
	Repo       *repository.AchievementRepository
	Mongo      *mongo.Client
}

func NewAchievementService(repo *repository.AchievementRepository, mongo *mongo.Client) *AchievementService {
	return &AchievementService{
		Repo:  repo,
		Mongo: mongo,
	}
}

// Input dari mahasiswa
type AchievementInput struct {
	AchievementType string                 `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
	Points          float64                `json:"points"`
}

// Output setelah submit
type AchievementOutput struct {
	MongoID   string                 `json:"mongo_id"`
	StudentID string                 `json:"student_id"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
}

func (s *AchievementService) CreateAchievement(
	ctx context.Context,
	userID int64,
	role string,
	input AchievementInput,
) (*AchievementOutput, error) {

	// 1. Pastikan role adalah mahasiswa
	if role != "mahasiswa" {
		return nil, errors.New("only students can create achievements")
	}

	// 2. Ambil student_id dari PostgreSQL
	studentID, err := s.Repo.GetStudentID(ctx, userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// 3. Siapkan dokumen MongoDB
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

	// 4. Insert ke MongoDB
	collection := s.Mongo.Database("uas").Collection("achievements")

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, errors.New("failed to save achievement to mongo")
	}

	objectID := result.InsertedID.(primitive.ObjectID).Hex()

	// 5. Insert reference ke PostgreSQL
	err = s.Repo.InsertReference(ctx, studentID, objectID)
	if err != nil {
		return nil, errors.New("failed to save reference to postgres")
	}

	// 6. Response output
	return &AchievementOutput{
		MongoID:   objectID,
		StudentID: studentID,
		Status:    "draft",
		Data:      doc,
	}, nil
}
