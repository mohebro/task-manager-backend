package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mohebro/task-manager-backend/internal/model"
	"github.com/mohebro/task-manager-backend/internal/repository"
	"github.com/mohebro/task-manager-backend/internal/service"
)

var errBoom = errors.New("db exploded")

// fakeAuthService is a controllable authService: each test sets what
// Register/Login/ParseToken return, so handler tests need no bcrypt,
// no JWT and no database.
type fakeAuthService struct {
	registerUser  *model.User
	registerToken string
	registerErr   error
	loginUser     *model.User
	loginToken    string
	loginErr      error
	parseID       uuid.UUID
	parseErr      error
}

func (f *fakeAuthService) Register(_ context.Context, _ model.RegisterInput) (*model.User, string, error) {
	return f.registerUser, f.registerToken, f.registerErr
}

func (f *fakeAuthService) Login(_ context.Context, _ model.LoginInput) (*model.User, string, error) {
	return f.loginUser, f.loginToken, f.loginErr
}

func (f *fakeAuthService) ParseToken(_ string) (uuid.UUID, error) {
	return f.parseID, f.parseErr
}

// fakeUserRepo is an in-memory UserRepository for handler tests.
type fakeUserRepo struct {
	byID map[uuid.UUID]*model.User
}

func (f *fakeUserRepo) Create(_ context.Context, user *model.User) (*model.User, error) {
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, _ string) (*model.User, error) {
	return nil, repository.ErrUserNotFound
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	user, ok := f.byID[id]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func testUser() *model.User {
	return &model.User{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "Mohsen",
		Email:     "mohsen@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// newTestEngine wires the real router shape (RequestLogger +
// AuthMiddleware + handler) around fakes.
func newTestEngine(t *testing.T, svc *fakeAuthService, repo *fakeUserRepo, dev bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(svc, repo, discardLogger(), dev)
	r := gin.New()
	r.Use(RequestLogger(discardLogger()))
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.GET("/me", AuthMiddleware(h.ParseToken), h.Me)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegister(t *testing.T) {
	user := testUser()

	t.Run("success returns 201 with user and token, no password hash", func(t *testing.T) {
		svc := &fakeAuthService{registerUser: user, registerToken: "tok-123"}
		r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, false)
		w := doRequest(t, r, http.MethodPost, "/auth/register",
			`{"name":"Mohsen","email":"mohsen@example.com","password":"password123"}`, "")
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201 (body %s)", w.Code, w.Body.String())
		}
		var resp model.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Token != "tok-123" || resp.User.Email != "mohsen@example.com" {
			t.Fatalf("unexpected body: %+v", resp)
		}
		if strings.Contains(w.Body.String(), "password_hash") {
			t.Fatal("response leaks password_hash")
		}
	})

	t.Run("invalid json returns 400", func(t *testing.T) {
		svc := &fakeAuthService{registerUser: user, registerToken: "tok-123"}
		r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, false)
		w := doRequest(t, r, http.MethodPost, "/auth/register", `{not json`, "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
	})

	t.Run("taken email returns 409", func(t *testing.T) {
		svc := &fakeAuthService{registerErr: repository.ErrEmailTaken}
		r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, false)
		w := doRequest(t, r, http.MethodPost, "/auth/register",
			`{"name":"Mohsen","email":"mohsen@example.com","password":"password123"}`, "")
		if w.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", w.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	user := testUser()

	t.Run("success returns 200 with token", func(t *testing.T) {
		svc := &fakeAuthService{loginUser: user, loginToken: "tok-456"}
		r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, false)
		w := doRequest(t, r, http.MethodPost, "/auth/login",
			`{"email":"mohsen@example.com","password":"password123"}`, "")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
		}
	})

	t.Run("wrong credentials return 401", func(t *testing.T) {
		svc := &fakeAuthService{loginErr: service.ErrInvalidCredentials}
		r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, false)
		w := doRequest(t, r, http.MethodPost, "/auth/login",
			`{"email":"mohsen@example.com","password":"wrong"}`, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})
}

func TestMe(t *testing.T) {
	user := testUser()
	repo := &fakeUserRepo{byID: map[uuid.UUID]*model.User{user.ID: user}}

	t.Run("good token returns the user", func(t *testing.T) {
		svc := &fakeAuthService{parseID: user.ID}
		r := newTestEngine(t, svc, repo, false)
		w := doRequest(t, r, http.MethodGet, "/auth/me", "", "good-token")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
		}
		var got model.PublicUser
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ID != user.ID {
			t.Fatalf("id = %v, want %v", got.ID, user.ID)
		}
	})

	t.Run("no token returns 401", func(t *testing.T) {
		svc := &fakeAuthService{parseID: user.ID}
		r := newTestEngine(t, svc, repo, false)
		w := doRequest(t, r, http.MethodGet, "/auth/me", "", "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	})

	t.Run("unknown user returns 404", func(t *testing.T) {
		svc := &fakeAuthService{parseID: uuid.MustParse("22222222-2222-2222-2222-222222222222")}
		r := newTestEngine(t, svc, repo, false)
		w := doRequest(t, r, http.MethodGet, "/auth/me", "", "good-token")
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})
}

func TestInternalErrorHidesDetailInProdShowsItInDev(t *testing.T) {
	// An unexpected service error must be logged (discarded here) and
	// answered 500. The client sees the exact error only in dev.
	for _, dev := range []bool{false, true} {
		t.Run(map[bool]string{false: "prod", true: "dev"}[dev], func(t *testing.T) {
			svc := &fakeAuthService{registerErr: errBoom}
			r := newTestEngine(t, svc, &fakeUserRepo{byID: map[uuid.UUID]*model.User{}}, dev)
			w := doRequest(t, r, http.MethodPost, "/auth/register",
				`{"name":"Mohsen","email":"mohsen@example.com","password":"password123"}`, "")
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", w.Code)
			}
			body := w.Body.String()
			if !strings.Contains(body, "request_id") {
				t.Fatalf("body has no request_id: %s", body)
			}
			if dev && !strings.Contains(body, "db exploded") {
				t.Fatalf("dev should show the exact error: %s", body)
			}
			if !dev && strings.Contains(body, "db exploded") {
				t.Fatalf("prod must not leak the exact error: %s", body)
			}
		})
	}
}
