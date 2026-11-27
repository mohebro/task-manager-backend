package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mohebro/task-manager-backend/internal/model"
	"github.com/mohebro/task-manager-backend/internal/repository"
)

var (
	ErrInvalidName        = errors.New("name is required")
	ErrInvalidEmail       = errors.New("a valid email is required")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// AuthService owns register, login and token signing.
// It depends on the UserRepository interface, not on Postgres,
// so it can be tested with a fake repository.
type AuthService struct {
	repo   repository.UserRepository
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewAuthService(repo repository.UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{repo: repo, secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (s *AuthService) Register(ctx context.Context, input model.RegisterInput) (*model.User, string, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, "", ErrInvalidName
	}
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, "", err
	}
	if len(input.Password) < 8 {
		return nil, "", ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	user, err := s.repo.Create(ctx, &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, "", err
	}
	token, err := s.signToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (*model.User, string, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Do not reveal whether the email exists.
		return nil, "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return nil, "", ErrInvalidCredentials
	}
	token, err := s.signToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// ParseToken validates a JWT and returns the user id it carries.
func (s *AuthService) ParseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidCredentials
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidCredentials
	}
	sub, ok := (*claims)["sub"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidCredentials
	}
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}
	return id, nil
}

func (s *AuthService) signToken(userID uuid.UUID) (string, error) {
	now := s.now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(s.ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return "", ErrInvalidEmail
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", ErrInvalidEmail
	}
	return email, nil
}
