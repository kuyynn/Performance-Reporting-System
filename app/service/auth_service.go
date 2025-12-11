package service

import (
	"errors"
	"time"

	"uas/app/model"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token       string      `json:"token"`
	User        *model.User `json:"user"`
	Permissions []string    `json:"permissions"`
}

func (s *AuthService) Login(input LoginInput) (*LoginOutput, error) {

	// 1. cek user
	user, err := s.userRepo.FindByUsernameOrEmail(input.Username)
	if err != nil {
		return nil, errors.New("invalid_credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user_inactive")
	}

	// 2. cek password
	if !utils.CheckPassword(user.PasswordHash, input.Password) {
		return nil, errors.New("invalid_credentials")
	}

	// 3. ambil permissions
	perms, err := s.userRepo.GetPermissionsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	// 4. generate JWT
	secret := []byte("RAHASIA_TOKEN_APLIKASI")
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Role,
		secret,
	)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{
		Token:       token,
		User:        user,
		Permissions: perms,
	}, nil
}

// LOGIN HANDLER
func (s *AuthService) LoginHandler(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	output, err := s.Login(input)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// generate refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 hari
	refreshToken := uuid.New().String()

	// simpan ke PostgreSQL
	err = s.tokenRepo.Save(c.Context(), output.User.ID, refreshToken, expiresAt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed_generate_refresh"})
	}
	return c.JSON(fiber.Map{
		"access_token":  output.Token,
		"refresh_token": refreshToken,
		"user":          output.User,
		"permissions":   output.Permissions,
	})
}

// REFRESH TOKEN
func (s *AuthService) RefreshHandler(c *fiber.Ctx) error {
	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_request"})
	}

	rt, err := s.tokenRepo.Get(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid_refresh_token"})
	}
	if time.Now().After(rt.ExpiresAt) {
		return c.Status(401).JSON(fiber.Map{"error": "refresh_expired"})
	}

	// get user
	user, err := s.userRepo.FindById(c.Context(), rt.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user_not_found"})
	}

	// buat token baru
	secret := []byte("RAHASIA_TOKEN_APLIKASI")
	newToken, err := utils.GenerateToken(user.ID, user.Username, user.Role, secret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token_failed"})
	}
	return c.JSON(fiber.Map{
		"access_token":  newToken,
		"refresh_token": req.RefreshToken,
	})
}

// PROFILE
func (s *AuthService) ProfileHandler(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	user, err := s.userRepo.FindById(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user_not_found"})
	}
	return c.JSON(user)
}

// LOGOUT
func (s *AuthService) LogoutHandler(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.Claims)
	err := s.tokenRepo.DeleteByUserID(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "logout_failed"})
	}
	return c.JSON(fiber.Map{"message": "logged_out"})
}
