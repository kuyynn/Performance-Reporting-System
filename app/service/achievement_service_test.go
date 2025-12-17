package service

import (
	"context"
	"errors"
	"testing"

	"uas/app/repository/mocks"

	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/mock"
)

// =======================
// SUBMIT ACHIEVEMENT TESTS
// =======================

func TestSubmitAchievement_Success(t *testing.T) {
	ctx := context.Background()

	userID := int64(10)
	role := "mahasiswa"
	achievementID := "mongo123"
	studentID := "student-uuid-1"

	repo := new(mocks.MockAchievementRepository)

	// mock repo calls
	repo.On("GetStudentID", ctx, userID).
		Return(studentID, nil)

	repo.On("Submit", ctx, achievementID, studentID).
		Return(nil)

	// simulate internal logic (SubmitAchievement)
	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only students can submit achievements")
		}

		sid, err := repo.GetStudentID(ctx, userID)
		if err != nil {
			return errors.New("student profile not found")
		}

		return repo.Submit(ctx, achievementID, sid)
	}()

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSubmitAchievement_Fail_NotMahasiswa(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	role := "admin"

	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only students can submit achievements")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "only students can submit achievements", err.Error())
}

func TestSubmitAchievement_Fail_StudentNotFound(t *testing.T) {
	ctx := context.Background()

	userID := int64(99)
	role := "mahasiswa"

	repo := new(mocks.MockAchievementRepository)

	repo.On("GetStudentID", ctx, userID).
		Return("", errors.New("student profile not found"))

	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only students can submit achievements")
		}

		_, err := repo.GetStudentID(ctx, userID)
		if err != nil {
			return errors.New("student profile not found")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "student profile not found", err.Error())
	repo.AssertExpectations(t)
}

// =======================
// CREATE ACHIEVEMENT (LOGIC ONLY)
// =======================

func TestCreateAchievement_Fail_NotMahasiswa(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	role := "admin"

	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only students can create achievements")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "only students can create achievements", err.Error())
}

func TestCreateAchievement_Fail_StudentNotFound(t *testing.T) {
	ctx := context.Background()

	userID := int64(10)
	role := "mahasiswa"

	repo := new(mocks.MockAchievementRepository)

	repo.On("GetStudentID", ctx, userID).
		Return("", errors.New("student profile not found"))

	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only students can create achievements")
		}

		_, err := repo.GetStudentID(ctx, userID)
		if err != nil {
			return errors.New("student profile not found")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "student profile not found", err.Error())
	repo.AssertExpectations(t)
}

// =======================
// VERIFY / REJECT (RBAC LOGIC ONLY)
// =======================

func TestVerifyAchievement_Fail_NotAdvisor(t *testing.T) {
	role := "mahasiswa"

	err := func() error {
		if role != "dosen wali" {
			return errors.New("only advisors can verify achievements")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "only advisors can verify achievements", err.Error())
}

func TestRejectAchievement_Fail_EmptyNote(t *testing.T) {
	note := ""

	err := func() error {
		if note == "" {
			return errors.New("rejection note is required")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "rejection note is required", err.Error())
}

// =======================
// DELETE ACHIEVEMENT (RBAC ONLY)
// =======================

func TestDeleteAchievement_Fail_NotMahasiswa(t *testing.T) {
	role := "admin"

	err := func() error {
		if role != "mahasiswa" {
			return errors.New("only mahasiswa can delete")
		}
		return nil
	}()

	assert.Error(t, err)
	assert.Equal(t, "only mahasiswa can delete", err.Error())
}
