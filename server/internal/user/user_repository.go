package user

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	var insertId int
	query := "INSERT INTO users(username, password, email) VALUES ($1, $2, $3) returning id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&insertId)
	if err != nil {
		return nil, err
	}

	user.ID = int64(insertId)
	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	query := "SELECT id, email, username, password FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Username, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) CreateSession(ctx context.Context, session *Session) (*Session, error) {
	query := `INSERT INTO sessions (id, email, refresh_token, is_revoked, created_at, expires_at)
			  VALUES($1, $2, $3, $4, $5, $6) RETURNING *`
	var s Session
	err := r.db.QueryRowContext(ctx,
		query,
		session.ID,
		session.Email,
		session.RefreshToken,
		session.IsRevoked,
		session.CreatedAt,
		session.ExpiresAt,
	).Scan(&s.ID, &s.Email, &s.RefreshToken, &s.IsRevoked, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting sessions: %w", err)
	}
	return &s, nil
}
func (r *repository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	query := `SELECT id, email, refresh_token, is_revoked, created_at, expires_at 
              FROM sessions WHERE refresh_token = $1 AND is_revoked = false`
	var s Session
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&s.ID, &s.Email, &s.RefreshToken, &s.IsRevoked, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("error failed to retrieve sessions: %w", err)
	}

	return &s, nil
}

func (r *repository) RevokeSession(ctx context.Context, refreshToken string) error {
	query := `UPDATE sessions SET is_revoked = true WHERE refresh_token = $1`
	_, err := r.db.ExecContext(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("error failed to revoke sessions: %w", err)
	}
	return nil
}

func (r *repository) DeleteSession(ctx context.Context, refreshToken string) error {
	query := `DELETE FROM sessions WHERE refresh_token = $1`
	_, err := r.db.ExecContext(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("error failed to delete sessions: %w", err)
	}
	return nil
}
