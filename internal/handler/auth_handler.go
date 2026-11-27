package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mohebro/task-manager-backend/internal/model"
	"github.com/mohebro/task-manager-backend/internal/repository"
	"github.com/mohebro/task-manager-backend/internal/service"
)

// ErrorResponse is the shape of every error body: {"error": "..."}.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid email or password"`
}

type authService interface {
	Register(ctx context.Context, input model.RegisterInput) (*model.User, string, error)
	Login(ctx context.Context, input model.LoginInput) (*model.User, string, error)
	ParseToken(tokenString string) (uuid.UUID, error)
}

type AuthHandler struct {
	service authService
	users   repository.UserRepository
	logger  *slog.Logger
	// dev is true only when APP_ENV=dev. It decides whether a 500
	// returns the exact error to the client. Never log or return
	// passwords or tokens; the error here never contains them.
	dev bool
}

func NewAuthHandler(service authService, users repository.UserRepository, logger *slog.Logger, dev bool) *AuthHandler {
	return &AuthHandler{service: service, users: users, logger: logger, dev: dev}
}

// ParseToken delegates to the auth service, so the router can
// pass AuthHandler.ParseToken straight to AuthMiddleware.
func (h *AuthHandler) ParseToken(tokenString string) (uuid.UUID, error) {
	return h.service.ParseToken(tokenString)
}

// Register creates a user and returns them with a JWT.
// 201 on success, 400 on invalid input, 409 when the email is taken.
//
// @Summary Register a new user
// @Description Creates a user with a bcrypt-hashed password and returns a JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body model.RegisterInput true "Registration details"
// @Success 201 {object} model.AuthResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 409 {object} handler.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input model.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, token, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, model.AuthResponse{User: user.Public(), Token: token})
}

// Login returns a JWT for a registered user.
// 401 on wrong email or password, without saying which one.
//
// @Summary Log in
// @Description Returns a JWT for a registered user. Wrong email or password both give 401.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body model.LoginInput true "Login details"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} handler.ErrorResponse
// @Failure 401 {object} handler.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input model.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, token, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.AuthResponse{User: user.Public(), Token: token})
}

// Me returns the user behind the Bearer token on this request.
//
// @Summary Get the current user
// @Description Returns the user behind the Bearer token.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.PublicUser
// @Failure 401 {object} handler.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, user.Public())
}

func (h *AuthHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, service.ErrInvalidName),
		errors.Is(err, service.ErrInvalidEmail),
		errors.Is(err, service.ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		// The exact error is always logged server-side with the
		// request id, so a later failure can be traced. The client
		// only sees the detail in dev; in production it gets a
		// generic message and the request id to report.
		requestID, _ := RequestIDFromContext(c)
		h.logger.Error("internal error",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"error", err.Error(),
		)
		if h.dev {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "request_id": requestID})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "request_id": requestID})
	}
}
