package service

import (
	"uas/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportService struct {
	AchievementRepo      *repository.AchievementRepository
	MongoAchievementRepo *repository.MongoAchievementRepository
}

func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	ctx := c.Context()

	// ================================
	// 1️⃣ Ambil achievement VERIFIED dari PostgreSQL
	// ================================
	refs, err := s.AchievementRepo.GetVerifiedAchievementRefs(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_verified_refs",
		})
	}

	// ================================
	// 2️⃣ Kumpulkan Mongo ObjectID
	// ================================
	var mongoIDs []primitive.ObjectID
	for _, r := range refs {
		idHex := r["mongo_id"].(string)
		objID, err := primitive.ObjectIDFromHex(idHex)
		if err != nil {
			continue
		}
		mongoIDs = append(mongoIDs, objID)
	}

	if len(mongoIDs) == 0 {
		return c.JSON(fiber.Map{
			"total_by_type":         map[string]int{},
			"total_by_year":         map[int]int{},
			"distribution_by_level": map[string]int{},
			"top_students":          []fiber.Map{},
		})
	}

	// ================================
	// 3️⃣ Ambil DETAIL achievement dari MongoDB
	// ================================
	docs, err := s.MongoAchievementRepo.FindByIDs(ctx, mongoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_mongo_achievements",
		})
	}

	// ================================
	// 4️⃣ AGREGASI (PAKAI bson.M)
	// ================================
	totalByType := map[string]int{}
	totalByYear := map[int]int{}
	distributionByLevel := map[string]int{}

	for _, d := range docs {
	totalByType[d.AchievementType]++
	totalByYear[d.Details.Year]++
	distributionByLevel[d.Details.CompetitionLevel]++
	}

	// ================================
	// 5️⃣ TOP MAHASISWA (POSTGRES ONLY)
	// ================================
	topStudents, err := s.AchievementRepo.GetTopStudents(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_top_students",
		})
	}

	// ================================
	// 6️⃣ RESPONSE FINAL FR-011
	// ================================
	return c.JSON(fiber.Map{
		"total_by_type":         totalByType,
		"total_by_year":         totalByYear,
		"distribution_by_level": distributionByLevel,
		"top_students":          topStudents,
	})
}

// ADMIN: GET /reports/student/:id
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	ctx := c.Context()
	studentID := c.Params("id")

	// 1️⃣ Ambil references dari PostgreSQL
	refs, err := s.AchievementRepo.GetByStudentUUID(ctx, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_student_refs",
		})
	}

	if len(refs) == 0 {
		return c.JSON(fiber.Map{
			"student_id":         studentID,
			"total_achievements": 0,
			"by_type":            map[string]int{},
			"by_year":            map[int]int{},
			"achievements":       []fiber.Map{},
		})
	}

	// 2️⃣ Ambil Mongo IDs
	var mongoIDs []primitive.ObjectID
	for _, r := range refs {
		idHex := r["mongo_id"].(string)
		objID, err := primitive.ObjectIDFromHex(idHex)
		if err != nil {
			continue
		}
		mongoIDs = append(mongoIDs, objID)
	}

	// 3️⃣ Ambil detail Mongo
	docs, err := s.MongoAchievementRepo.FindByIDs(ctx, mongoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed_get_mongo_details",
		})
	}

	// 4️⃣ Agregasi
	byType := map[string]int{}
	byYear := map[int]int{}
	var achievements []fiber.Map

	for _, d := range docs {
    // by type
    byType[d.AchievementType]++
    // by year
    byYear[d.Details.Year]++
    achievements = append(achievements, fiber.Map{
        "type":  d.AchievementType,
        "year":  d.Details.Year,
        "level": d.Details.CompetitionLevel,
    })
}

	return c.JSON(fiber.Map{
		"student_id":         studentID,
		"total_achievements": len(achievements),
		"by_type":            byType,
		"by_year":            byYear,
		"achievements":       achievements,
	})
}
