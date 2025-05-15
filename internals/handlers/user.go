package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/assaidy/iWonder/internals/db"
	"github.com/assaidy/iWonder/internals/repository"
	"github.com/assaidy/iWonder/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserPayload struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio,omitempty"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,customNoOuterSpaces,max=100"`
	Bio      string `json:"bio" validate:"customNoOuterSpaces"`
	Username string `json:"username" validate:"required,customUsername,max=50"`
	Password string `json:"password" validate:"required,customNoOuterSpaces,max=50"`
}

func HandleRegister(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("error hashing password :%v", err)
	}

	affectedRows, err := queries.InsertUser(context.Background(), repository.InsertUserParams{
		ID:             uuid.New(),
		Name:           req.Name,
		Bio:            sql.NullString{String: req.Bio, Valid: req.Bio != ""},
		Username:       req.Username,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return fmt.Errorf("error inserting user: %v", err)
	}
	if affectedRows == 0 {
		return c.Status(fiber.StatusConflict).SendString("username already exists")
	}

	return c.Status(fiber.StatusCreated).SendString("user created successfully")
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,customUsername,max=50"`
	Password string `json:"password" validate:"required,customNoOuterSpaces,max=50"`
}

func HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	repoUser, err := queries.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).SendString("user not found")
		}
		return fmt.Errorf("error getting user: %v", err)
	}

	if !utils.VerifyPassword(req.Password, repoUser.HashedPassword) {
		return c.Status(fiber.StatusUnauthorized).SendString("invalid password")
	}

	accessToken, err := utils.GenerateJWTAccessToken(utils.JwtClaims{
		UserID: repoUser.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpirationMinutes * time.Minute)),
		},
	})
	if err != nil {
		return fmt.Errorf("error creating jwt access token: %v", err)
	}

	refreshToken := utils.GenerateRefreshToken()
	if err := queries.InsertRefreshToken(context.Background(), repository.InsertRefreshTokenParams{
		Token:     refreshToken,
		UserID:    repoUser.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * RefreshTokenExpirationDays),
	}); err != nil {
		return fmt.Errorf("error creating refresh token: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": UserPayload{
			ID:        repoUser.ID,
			Name:      repoUser.Name,
			Bio:       repoUser.Bio.String,
			Username:  repoUser.Username,
			CreatedAt: repoUser.CreatedAt,
		},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

type GetAccessTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func HandleGetAccessToken(c *fiber.Ctx) error {
	var req GetAccessTokenRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	repoRefreshToken, err := queries.GetRefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).SendString("refresh token not found")
		}
		return fmt.Errorf("error getting refresh token: %v", err)
	}

	if repoRefreshToken.ExpiresAt.Sub(time.Now()) < 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("token expired")
	}

	accessToken, err := utils.GenerateJWTAccessToken(utils.JwtClaims{
		UserID: repoRefreshToken.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpirationMinutes * time.Minute)),
		},
	})
	if err != nil {
		return fmt.Errorf("error creating jwt access token: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken": accessToken,
	})
}

func HandleGetUserByID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid user id")
	}

	repoUser, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).SendString("user not found")
		}
		return fmt.Errorf("error getting user: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": UserPayload{
			ID:        repoUser.ID,
			Name:      repoUser.Name,
			Bio:       repoUser.Bio.String,
			Username:  repoUser.Username,
			CreatedAt: repoUser.CreatedAt,
		},
	})
}

func HandleGetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("user_id")

	repoUser, err := queries.GetUserByUsername(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).SendString("user not found")
		}
		return fmt.Errorf("error getting user: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": UserPayload{
			ID:        repoUser.ID,
			Name:      repoUser.Name,
			Bio:       repoUser.Bio.String,
			Username:  repoUser.Username,
			CreatedAt: repoUser.CreatedAt,
		},
	})
}

type UpdateRequest struct {
	Name     string `json:"name" validate:"required,customNoOuterSpaces,max=100"`
	Bio      string `json:"bio" validate:"customNoOuterSpaces"`
	Username string `json:"username" validate:"required,customUsername,max=50"`
	Password string `json:"password" validate:"required,customNoOuterSpaces,max=50"`
}

func HandleUpdateUser(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	userID := getAuthedUserID(c)

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("err begin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckUsername(context.Background(), req.Username); err != nil {
		return fmt.Errorf("error checking username: %v", err)
	} else if ok {
		return c.Status(fiber.StatusConflict).SendString("username already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("error hashing password :%v", err)
	}

	err = qtx.UpdateUserByID(context.Background(), repository.UpdateUserByIDParams{
		ID:             userID,
		Name:           req.Name,
		Bio:            sql.NullString{String: req.Bio, Valid: req.Bio != ""},
		Username:       req.Username,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx :%v", err)
	}

	return c.Status(fiber.StatusOK).SendString("user updated successfully")
}

func HandleDeleteUser(c *fiber.Ctx) error {
	userID := getAuthedUserID(c)
	if err := queries.DeleteUserById(context.Background(), userID); err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
