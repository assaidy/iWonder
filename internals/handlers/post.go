package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/assaidy/iWonder/internals/db"
	"github.com/assaidy/iWonder/internals/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PostPayload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userID"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Answered  bool      `json:"answered"`
}

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,customNoOuterSpaces,max=200"`
	Content string `json:"content" validate:"required,customNoOuterSpaces"`
}

func HandleCreatePost(c *fiber.Ctx) error {
	var req CreatePostRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	userID := getAuthedUserID(c)

	repoPost, err := queries.InsertPost(context.Background(), repository.InsertPostParams{
		ID:      uuid.New(),
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return fmt.Errorf("error inserting post: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"post": PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		},
	})
}

func HandleGetPostByID(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}

	repoPost, err := queries.GetPostByID(context.Background(), postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).SendString("post not found")
		}
		return fmt.Errorf("error getting post: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"post": PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		},
	})
}

type UpdatePostRequest struct {
	Title   string `json:"title" validate:"required,customNoOuterSpaces,max=200"`
	Content string `json:"content" validate:"required,customNoOuterSpaces"`
}

func HandleUpdatePostByID(c *fiber.Ctx) error {
	var req UpdatePostRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}
	userID := getAuthedUserID(c)

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckPostForUser(context.Background(), repository.CheckPostForUserParams{
		ID:     postID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found for user")
	}

	repoPost, err := qtx.UpdatePostByID(context.Background(), repository.UpdatePostByIDParams{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return fmt.Errorf("error updating post: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"post": PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		},
	})
}

func HandleTogglePostAnswered(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}
	userID := getAuthedUserID(c)

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckPostForUser(context.Background(), repository.CheckPostForUserParams{
		ID:     postID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found for user")
	}

	repoPost, err := qtx.TogglePostAnswered(context.Background(), postID)
	if err != nil {
		return fmt.Errorf("error toggle post answered: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"post": PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		},
	})
}

func HandleDeletePostByID(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}
	userID := getAuthedUserID(c)

	if ok, err := queries.CheckPostForUser(context.Background(), repository.CheckPostForUserParams{
		ID:     postID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found for user")
	}

	if err := queries.DeletePostByID(context.Background(), postID); err != nil {
		return fmt.Errorf("error deleting post: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("post deleted successfully")
}
