package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/assaidy/iWonder/internals/db"
	"github.com/assaidy/iWonder/internals/repository"
	"github.com/assaidy/iWonder/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	AccessTokenExpirationMinutes = 10
	RefreshTokenExpirationDays   = 7
	AuthedUserID                 = "middleware.auth.userID"
)

var (
	queries = repository.New(db.Connection)
)

func parseAndValidateJsonBody(c *fiber.Ctx, out any) error {
	if err := c.BodyParser(out); err != nil {
		return fmt.Errorf("invalid json body")
	}
	if err := utils.ValidateStruct(out); err != nil {
		return fmt.Errorf("invalid request data: %w", err)
	}
	return nil
}

func WithJwt(c *fiber.Ctx) error {
	tokenString := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
	if tokenString == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing or malformed Authorization header")
	}
	claims, err := utils.ParseJWTTokenString(tokenString)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	if claims.ExpiresAt.Sub(time.Now()) < 0 {
		return fiber.ErrUnauthorized
	}
	// NOTE: if the users deleted his account, but his access token hasn't expired yet,
	// and we got a request that uses mwAuth(get's userid from context),
	// we need to ensure that user exists.
	if exists, err := queries.CheckUserID(context.Background(), claims.UserID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error checking user ID: %+v", err))
	} else if !exists {
		return fiber.ErrUnauthorized
	}
	c.Locals(AuthedUserID, claims.UserID)
	return c.Next()
}

func getAuthedUserID(c *fiber.Ctx) uuid.UUID {
	return c.Locals(AuthedUserID).(uuid.UUID)
}
