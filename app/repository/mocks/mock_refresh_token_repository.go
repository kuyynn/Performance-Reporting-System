package mocks

import (
	"context"
	"time"

	"uas/app/model"

	"github.com/stretchr/testify/mock"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Save(
	ctx context.Context,
	userID int64,
	token string,
	expiresAt time.Time,
) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Get(
	ctx context.Context,
	token string,
) (*model.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByUserID(
	ctx context.Context,
	userID int64,
) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
