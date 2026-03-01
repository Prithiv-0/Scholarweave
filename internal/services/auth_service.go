package services

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"Scholarweave/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthService struct {
	pool *pgxpool.Pool
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) Register(name, email, password string) (*models.User, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)

	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return nil, errors.New("invalid email format")
	}
	if password == "" || len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return nil, errors.New("password must be at most 72 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	_, err = s.pool.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, name, email, string(hash), now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("email already registered")
		}
		slog.Error("failed to insert user", "error", err)
		return nil, errors.New("failed to create user")
	}

	return &models.User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: now,
	}, nil
}

func (s *AuthService) Login(email, password string) (*models.User, error) {
	email = normalizeEmail(email)

	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	var user models.User
	var passwordHash string

	err := s.pool.QueryRow(context.Background(),
		`SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &passwordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		slog.Error("failed to query user", "error", err)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *AuthService) GetByID(id string) (*models.User, error) {
	var user models.User

	err := s.pool.QueryRow(context.Background(),
		`SELECT id, name, email, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		slog.Error("failed to query user by id", "error", err)
		return nil, errors.New("user not found")
	}

	return &user, nil
}
