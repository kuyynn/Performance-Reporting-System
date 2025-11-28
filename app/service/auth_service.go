package service

import (
	"context"
	"errors"
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Token       string
	User        *model.User
	Permissions []model.Permission
}

func (s *AuthService) Login(input LoginInput) (*LoginOutput, error) {

	// cek user berdasarkan username / email
	user, err := s.userRepo.FindByUsernameOrEmail(input.Username)
	if err != nil {
		return nil, errors.New("invalid_credentials")
	}

	// cek user aktif apa nggak
	if !user.IsActive {
		return nil, errors.New("user_inactive")
	}

	// cek password cocok apa nggak
	if !utils.CheckPassword(user.PasswordHash, input.Password) {
		return nil, errors.New("invalid_credentials")
	}

	// ambil detail user dari tabel (biasanya buat token)
	userDetail, err := s.userRepo.FindById(context.Background(), user.ID)
	if err != nil {
		return nil, err
	}

	// ambil permissions user berdasarkan role
	perms, err := s.userRepo.GetPermissionsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	// secret key buat generate token
	secret := []byte("RAHASIA_TOKEN_APLIKASI") // ntar bisa disimpen di env biar aman

	// generate token
	token, _, err := utils.GenerateToken(*userDetail, secret)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:       token,
		User:        user,
		Permissions: perms,
	}, nil
}
