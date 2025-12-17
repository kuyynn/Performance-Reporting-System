package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"uas/app/model"
	"uas/app/repository/mocks"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =======================
// SETUP
// =======================

func setupAuthApp(service *AuthService) *fiber.App {
	app := fiber.New()

	// mock auth middleware (claims)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("claims", &utils.Claims{
			UserID: 1,
			Role:   "admin",
		})
		return c.Next()
	})

	app.Post("/login", service.LoginHandler)
	app.Post("/refresh", service.RefreshHandler)
	app.Get("/profile", service.ProfileHandler)
	app.Post("/logout", service.LogoutHandler)

	return app
}

// =======================
// LOGIN (LOGIC)
// =======================

func TestAuth_Login_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)

	service := NewAuthService(userRepo, tokenRepo)

	hash, _ := utils.HashPassword("secret")

	user := &model.User{
		ID:           1,
		Username:     "test",
		PasswordHash: hash,
		Role:         "admin",
		IsActive:     true,
	}

	userRepo.On("FindByUsernameOrEmail", "test").
		Return(user, nil)

	userRepo.On("GetPermissionsByUserID", int64(1)).
		Return([]string{"user.read"}, nil)

	output, err := service.Login(LoginInput{
		Username: "test",
		Password: "secret",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Token)
	assert.Equal(t, user.Username, output.User.Username)

	userRepo.AssertExpectations(t)
}

// =======================
// LOGIN HANDLER
// =======================

func TestAuth_LoginHandler_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)

	service := NewAuthService(userRepo, tokenRepo)
	app := setupAuthApp(service)

	hash, _ := utils.HashPassword("secret")

	user := &model.User{
		ID:           1,
		Username:     "test",
		PasswordHash: hash,
		Role:         "admin",
		IsActive:     true,
	}

	userRepo.On("FindByUsernameOrEmail", "test").
		Return(user, nil)

	userRepo.On("GetPermissionsByUserID", int64(1)).
		Return([]string{"user.read"}, nil)

	tokenRepo.On(
		"Save",
		mock.Anything,
		int64(1),
		mock.Anything,
		mock.Anything,
	).Return(nil)

	body, _ := json.Marshal(LoginInput{
		Username: "test",
		Password: "secret",
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// =======================
// REFRESH TOKEN
// =======================

func TestAuth_RefreshHandler_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)

	service := NewAuthService(userRepo, tokenRepo)
	app := setupAuthApp(service)

	rt := &model.RefreshToken{
		UserID:    1,
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tokenRepo.On(
		"Get",
		mock.Anything,
		"refresh-token",
	).Return(rt, nil)

	userRepo.On(
		"FindById",
		mock.Anything,
		int64(1),
	)	.Return(&model.UserResponse{
		ID:       1,
		Username: "test",
		Role:     "admin",
	}, nil)

	body, _ := json.Marshal(model.RefreshTokenRequest{
		RefreshToken: "refresh-token",
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/refresh",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// =======================
// LOGOUT
// =======================

func TestAuth_LogoutHandler_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)

	service := NewAuthService(userRepo, tokenRepo)
	app := setupAuthApp(service)

	tokenRepo.On(
		"DeleteByUserID",
		mock.Anything,
		int64(1),
	).Return(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/logout",
		nil,
	)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	tokenRepo.AssertExpectations(t)
}

// =======================
// LOGIN FAIL
// =======================

func TestAuth_Login_InvalidCredential(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)

	service := NewAuthService(userRepo, tokenRepo)

	userRepo.On(
		"FindByUsernameOrEmail",
		"test",
	).Return(nil, errors.New("not found"))

	_, err := service.Login(LoginInput{
		Username: "test",
		Password: "wrong",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid_credentials", err.Error())
}
