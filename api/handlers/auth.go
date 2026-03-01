package handlers

import (
	"strings"
	"time"

	"Scholarweave/internal/models"
	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService     *services.AuthService
	jwtSecret       []byte
	tokenExpiration time.Duration
}

func NewAuthHandler(authService *services.AuthService, jwtSecret string, tokenExpiration time.Duration) *AuthHandler {
	if tokenExpiration <= 0 {
		tokenExpiration = 24 * time.Hour
	}

	return &AuthHandler{
		authService:     authService,
		jwtSecret:       []byte(jwtSecret),
		tokenExpiration: tokenExpiration,
	}
}

func (h *AuthHandler) createToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(h.tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var request models.RegisterRequest
	if err := c.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	user, err := h.authService.Register(request.Name, request.Email, request.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.createToken(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate token")
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var request models.LoginRequest
	if err := c.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	user, err := h.authService.Login(request.Email, request.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	token, err := h.createToken(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	user, err := h.authService.GetByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(user)
}
