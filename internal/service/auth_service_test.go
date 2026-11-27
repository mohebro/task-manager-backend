package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mohebro/task-manager-backend/internal/model"
	"github.com/mohebro/task-manager-backend/internal/repository"
)

// fakeUserRepo is an in-memory UserRepository for service tests:
// no database is needed to test validation, hashing and tokens.
type fakeUserRepo struct {
	byEmail map[string]*model.User
	byID    map[uuid.UUID]*model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byEmail: map[string]*model.User{},
		byID:    map[uuid.UUID]*model.User{},
	}
}

func (f *fakeUserRepo) Create(_ context.Context, user *model.User) (*model.User, error) {
	if _, taken := f.byEmail[user.Email]; taken {
		return nil, repository.ErrEmailTaken
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	user, ok := f.byEmail[email]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	user, ok := f.byID[id]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "test-secret", time.Hour)

	user, token, err := svc.Register(context.Background(), model.RegisterInput{
		Name: "Mohsen", Email: "Mohsen@Example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != "mohsen@example.com" {
		t.Fatalf("email should be lowercased, got %q", user.Email)
	}
	if token == "" {
		t.Fatal("register should return a token")
	}

	id, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if id != user.ID {
		t.Fatalf("token user id = %v, want %v", id, user.ID)
	}

	// Same email, different case, is the same account.
	if _, _, err := svc.Register(context.Background(), model.RegisterInput{
		Name: "Other", Email: "MOHSEN@example.com", Password: "password123",
	}); !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("duplicate email should be ErrEmailTaken, got %v", err)
	}

	loggedIn, _, err := svc.Login(context.Background(), model.LoginInput{
		Email: "mohsen@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("login returned wrong user")
	}

	if _, _, err := svc.Login(context.Background(), model.LoginInput{
		Email: "mohsen@example.com", Password: "wrong-password",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password should be ErrInvalidCredentials, got %v", err)
	}
}

func TestRegisterValidation(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "test-secret", time.Hour)
	ctx := context.Background()

	cases := []struct {
		name  string
		input model.RegisterInput
		want  error
	}{
		{"empty name", model.RegisterInput{Name: "  ", Email: "a@b.com", Password: "password123"}, ErrInvalidName},
		{"bad email", model.RegisterInput{Name: "A", Email: "not-an-email", Password: "password123"}, ErrInvalidEmail},
		{"short password", model.RegisterInput{Name: "A", Email: "a@b.com", Password: "short"}, ErrInvalidPassword},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := svc.Register(ctx, tc.input); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
