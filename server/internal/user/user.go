package user

import (
	"context"
	"time"
)

type User struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type CreateUserResponse struct {
	ID       string `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
}

type LoginUserRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type LoginUserResponse struct {
	ID           string `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	Email        string `json:"email" db:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type Session struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	RefreshToken string    `db:"refresh_token"`
	IsRevoked    bool      `db:"is_revoked"`
	CreatedAt    time.Time `db:"created_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	RevokeSession(ctx context.Context, refreshToken string) error
	DeleteSession(ctx context.Context, refreshToken string) error
}

type Service interface {
	CreateUser(c context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
	Login(c context.Context, req *LoginUserRequest) (*LoginUserResponse, error)
	RefreshToken(c context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(c context.Context, req *LogoutRequest) error
}
