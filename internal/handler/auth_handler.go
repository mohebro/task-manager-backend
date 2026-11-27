package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mohebro/task-manager-backend/internal/model"
	"github.com/mohebro/task-manager-backend/internal/repository"
	"github.com/mohebro/task-manager-backend/internal/service"
)

type authService interface {
	Register(ctx context.Context, input model.RegisterInput) (*model.User, string, error)
	Login(ctx context.Context, input model.LoginInput) (*model.User, string, error)
	ParseToken(tokenString string) (uuid.UUID, error)
}

type AuthHandler struct {
	service authService
	users   repository.UserRepository
}

func NewAuthHandler(service authService, users repository.UserRepository) *AuthHandler {
	return &AuthHandler{service: service, users: users}
}

// ParseToken delegates to the auth service, so the router can
// pass AuthHandler.ParseToken straight to AuthMiddleware.
func (h *AuthHandler) ParseToken(tokenString string) (uuid.UUID, error) {
	return h.service.ParseToken(tokenString)
}

// Register creates a user and returns them with a JWT.
// 201 on success, 400 on invalid input, 409 when the email is taken.
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
