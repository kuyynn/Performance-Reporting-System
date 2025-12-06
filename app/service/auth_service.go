package service

import (
	"errors"
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token       string       `json:"token"`
	User        *model.User  `json:"user"`
	Permissions []string     `json:"permissions"`
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

	// 4. generate token
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

	// 5. return data
	return &LoginOutput{
		Token:       token,
		User:        user,
		Permissions: perms,
	}, nil
}
