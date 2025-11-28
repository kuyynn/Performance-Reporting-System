package utils

import (
	"time"
	"uas/app/model"
	"github.com/golang-jwt/jwt/v5"
)

type JWTCustomClaims struct {
	UserID      string             `json:"sub"`
	FullName    string             `json:"fullName"`
	Username    string             `json:"username"`
	Role        string             `json:"role"`
	Permissions []string           `json:"perms"`
	jwt.RegisteredClaims
}

func GenerateToken(User model.UserResponse, jwsScret []byte) (string, string, error) {
	var AccessExpiration = time.Now().Add(15 * time.Minute)
	AccessClaims := model.Claims{
		UserID:   User.ID,
		Username: User.Username,
		Role:     User.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(AccessExpiration),
		},
	}
	AccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims)
	accessString, err := AccessToken.SignedString(jwsScret)
	if err != nil {
		return "", "", err
	}
	RefreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	RefreshClaims := model.Claims{
		UserID: User.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(RefreshExpiration),
		},
	}
	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims)
	refreshString, err := RefreshToken.SignedString([]byte(jwsScret))

	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}