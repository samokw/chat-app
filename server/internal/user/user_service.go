package user

import (
	"context"
	"fmt"
	"server/internal/token"
	"server/internal/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type service struct {
	Repository
	timeout   time.Duration
	secretKey string
}

func NewService(repository Repository, secretKey string) Service {
	return &service{
		Repository: repository,
		timeout:    time.Duration(2) * time.Second,
		secretKey:  secretKey,
	}
}

func (s *service) CreateUser(c context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// TODO: hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}
	user, err := s.Repository.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}
	res := &CreateUserResponse{
		ID:       strconv.Itoa(int(user.ID)),
		Email:    user.Email,
		Username: user.Username,
	}
	return res, nil
}

func (s *service) Login(c context.Context, req *LoginUserRequest) (*LoginUserResponse, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	u, err := s.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	err = utils.CheckPassword(req.Password, u.Password)
	if err != nil {
		return nil, err
	}

	jwtMaker := token.NewJwtMaker(s.secretKey)
	accessToken, _, err := jwtMaker.CreateToken(int(u.ID), u.Email, time.Minute*15)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Create session
	sessionID := uuid.New().String()
	session := &Session{
		ID:           sessionID,
		Email:        u.Email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7), // 7 days
	}

	_, err = s.Repository.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Store tokens in context for handler to set cookies
	return &LoginUserResponse{
		ID:           strconv.Itoa(int(u.ID)),
		Username:     u.Username,
		Email:        u.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) RefreshToken(c context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Get session by refresh token
	session, err := s.Repository.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := s.Repository.GetUserByEmail(ctx, session.Email)
	if err != nil {
		return nil, err
	}

	jwtMaker := token.NewJwtMaker(s.secretKey)
	accessToken, _, err := jwtMaker.CreateToken(int(user.ID), user.Email, time.Minute*15)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		Message:     "Token refreshed successfully",
	}, nil
}

func (s *service) Logout(c context.Context, req *LogoutRequest) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	err := s.Repository.RevokeSession(ctx, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}
