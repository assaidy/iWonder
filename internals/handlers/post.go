package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func HandleGetPost(c *fiber.Ctx) error {
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

func HandleUpdatePost(c *fiber.Ctx) error {
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

	if err := qtx.UpdatePostByID(context.Background(), repository.UpdatePostByIDParams{
		Title:   req.Title,
		Content: req.Content,
	}); err != nil {
		return fmt.Errorf("error updating post: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("post updated successfully")
}

func HandleDeletePost(c *fiber.Ctx) error {
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

	return c.SendStatus(fiber.StatusNoContent)
}

type AddPostTagsRequest struct {
	Tags []string `json:"tags" validate:"required"`
}

func HandleAddPostTags(c *fiber.Ctx) error {
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

	var req AddPostTagsRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	for _, tag := range req.Tags {
		tagID, err := qtx.InsertTag(context.Background(), strings.ToLower(strings.TrimSpace(tag)))
		if err != nil {
			return fmt.Errorf("error inserting tag: %v", err)
		}

		if err := qtx.InsertTagForPost(context.Background(), repository.InsertTagForPostParams{
			PostID: postID,
			TagID:  tagID,
		}); err != nil {
			return fmt.Errorf("error inserting tag for post: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("err commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("tags added successfully")
}

func HandleGetPostTags(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}

	if ok, err := queries.CheckPost(context.Background(), postID); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found")
	}

	tags, err := queries.GetPostTags(context.Background(), postID)
	if err != nil {
		return fmt.Errorf("error getting post tags: %v", err)
	}
	if tags == nil {
		tags = []string{}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"tags": tags,
	})
}

func HandleDeletePostTag(c *fiber.Ctx) error {
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

	tagName := strings.ToLower(c.Params("tag_name"))

	if err := qtx.DeleteTagForPost(context.Background(), repository.DeleteTagForPostParams{
		PostID: postID,
		Name:   tagName,
	}); err != nil {
		return fmt.Errorf("error deleting post tag: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("err commit tx: %v", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,customNoOuterSpaces"`
}

func HandleCreateComment(c *fiber.Ctx) error {
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

	if ok, err := qtx.CheckPost(context.Background(), postID); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found")
	}

	var req CreateCommentRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := qtx.InsertComment(context.Background(), repository.InsertCommentParams{
		ID:      uuid.New(),
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	}); err != nil {
		return fmt.Errorf("error inserting comment: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusCreated).SendString("comment created successfully")
}

type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,customNoOuterSpaces"`
}

func HandleUpdateComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}
	userID := getAuthedUserID(c)

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckCommentForUser(context.Background(), repository.CheckCommentForUserParams{
		ID:     commentID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("error checking comment for user: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment not found for user")
	}

	var req UpdateCommentRequest
	if err := parseAndValidateJsonBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := qtx.UpdateComment(context.Background(), repository.UpdateCommentParams{
		ID:      commentID,
		Content: req.Content,
	}); err != nil {
		return fmt.Errorf("error udpating comment: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("comment updated successfully")
}

func HandleDeleteComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}
	userID := getAuthedUserID(c)

	if ok, err := queries.CheckCommentForUser(context.Background(), repository.CheckCommentForUserParams{
		ID:     commentID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("error checking comment for user: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment not found for user")
	}

	if err := queries.DeleteComment(context.Background(), commentID); err != nil {
		return fmt.Errorf("error deleting comment: %v", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type CommentsCursor struct {
	CreatedAt time.Time `json:"createdAt"`
}

type CommentPayload struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"postID"`
	UserID    uuid.UUID `json:"userID"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func HandleGetAllPostComments(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}

	if ok, err := queries.CheckPost(context.Background(), postID); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found")
	}

	limit := c.QueryInt("limit")
	if limit < 10 || limit > 100 {
		limit = 10
	}

	var requestCursor CommentsCursor
	if err := decodeBase64AndUnmarshalJson(&requestCursor, c.Query("cursor")); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid cursor format")
	}

	repoComments, err := queries.GetPostComments(context.Background(), repository.GetPostCommentsParams{
		PostID:    postID,
		CreatedAt: requestCursor.CreatedAt,
		Limit:     int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting post comments: %v", err)
	}

	var encodedResponseCursor string
	hasMore := limit < len(repoComments)
	if hasMore {
		responseCursor := CommentsCursor{
			CreatedAt: repoComments[limit].CreatedAt,
		}
		encodedResponseCursor, err = marshalJsonAndEncodeBase64(responseCursor)
		if err != nil {
			return fmt.Errorf("error encoding cursor: %v", err)
		}
		repoComments = repoComments[:limit]
	}

	comments := make([]CommentPayload, 0, len(repoComments))
	for _, repoComment := range repoComments {
		comments = append(comments, CommentPayload{
			ID:        repoComment.ID,
			PostID:    repoComment.PostID,
			UserID:    repoComment.UserID,
			Content:   repoComment.Content,
			CreatedAt: repoComment.CreatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"comments":   comments,
		"cursor":     encodedResponseCursor,
		"hasMore":    hasMore,
		"totalCount": len(comments),
	})
}

func HandleVoteComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	kind := c.Query("kind")
	if !(kind == "up" || kind == "down") {
		return c.Status(fiber.StatusBadRequest).SendString("invalid vote kind")
	}

	if ok, err := qtx.CheckComment(context.Background(), commentID); err != nil {
		return fmt.Errorf("error checking comment: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment not found")
	}

	userID := getAuthedUserID(c)

	if err := qtx.InsertCommentVote(context.Background(), repository.InsertCommentVoteParams{
		UserID:    userID,
		CommentID: commentID,
		Kind:      kind,
	}); err != nil {
		return fmt.Errorf("error inserting comment vote: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("vote made successfully")
}

func HandleUnvoteComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}

	userID := getAuthedUserID(c)

	if ok, err := queries.CheckCommentVoteForUser(context.Background(), repository.CheckCommentVoteForUserParams{
		CommentID: commentID,
		UserID:    userID,
	}); err != nil {
		return fmt.Errorf("error checking comment vote: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment vote not found for user")
	}

	if err := queries.DeleteCommentVote(context.Background(), repository.DeleteCommentVoteParams{
		CommentID: commentID,
		UserID:    userID,
	}); err != nil {
		return fmt.Errorf("error deleting comment vote: %v", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func HandleGetCommentVoteCounts(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckComment(context.Background(), commentID); err != nil {
		return fmt.Errorf("error checking comment: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment not found")
	}

	voteCounts, err := qtx.GetCommentVoteCounts(context.Background(), commentID)
	if err != nil {
		return fmt.Errorf("error getting comment vote counts: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"upCount":   voteCounts.UpCount,
		"downCount": voteCounts.DownCount,
	})
}

func HandleSetPostAnswer(c *fiber.Ctx) error {
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

	commentID, err := uuid.Parse(c.Query("commentID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid comment id")
	}

	if ok, err := qtx.CheckComment(context.Background(), commentID); err != nil {
		return fmt.Errorf("error checking comment: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("comment not found")
	}

	if err := qtx.InsertPostAnswer(context.Background(), repository.InsertPostAnswerParams{
		PostID:    postID,
		CommentID: commentID,
	}); err != nil {
		return fmt.Errorf("error setting post answer: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).SendString("post answerd successfully")
}

func HandleUnsetPostAnswer(c *fiber.Ctx) error {
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

	if err := queries.DeletePostAnswer(context.Background(), postID); err != nil {
		return fmt.Errorf("error deleting post answer: %v", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func HandleGetPostAnswer(c *fiber.Ctx) error {
	postID, err := uuid.Parse(c.Params("post_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid post id")
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return fmt.Errorf("error bigin tx: %v", err)
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	if ok, err := qtx.CheckPost(context.Background(), postID); err != nil {
		return fmt.Errorf("error checking post: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("post not found")
	}

	repoComment, err := qtx.GetPostAnswer(context.Background(), postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNoContent).SendString("no answer for this post")
		}
		return fmt.Errorf("error getting post answer: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit tx: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"comment": CommentPayload{
			ID:        repoComment.ID,
			PostID:    repoComment.PostID,
			UserID:    repoComment.UserID,
			Content:   repoComment.Content,
			CreatedAt: repoComment.CreatedAt,
		},
	})
}

type PostsCursor struct {
	CreatedAt time.Time `json:"createdAt"`
}

func HandleGetAllPostsForUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid user id")
	}

	if ok, err := queries.CheckUserID(context.Background(), userID); err != nil {
		return fmt.Errorf("error checking user id: %v", err)
	} else if !ok {
		return c.Status(fiber.StatusNotFound).SendString("user not found")
	}

	limit := c.QueryInt("limit")
	if limit < 10 || limit > 100 {
		limit = 10
	}

	var requestCursor PostsCursor
	if err := decodeBase64AndUnmarshalJson(&requestCursor, c.Query("cursor")); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid cursor format")
	}

	repoPosts, err := queries.GetUserPosts(context.Background(), repository.GetUserPostsParams{
		UserID:    userID,
		CreatedAt: requestCursor.CreatedAt,
		Limit:     int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting user posts: %v", err)
	}

	var encodedResponseCursor string
	hasMore := limit < len(repoPosts)
	if hasMore {
		responseCursor := PostsCursor{
			CreatedAt: repoPosts[limit].CreatedAt,
		}
		encodedResponseCursor, err = marshalJsonAndEncodeBase64(responseCursor)
		if err != nil {
			return fmt.Errorf("error encoding cursor: %v", err)
		}
		repoPosts = repoPosts[:limit]
	}

	posts := make([]PostPayload, 0, len(repoPosts))
	for _, repoPost := range repoPosts {
		posts = append(posts, PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"posts":      posts,
		"cursor":     encodedResponseCursor,
		"hasMore":    hasMore,
		"totalCount": len(posts),
	})
}

func HandleGetAllPosts(c *fiber.Ctx) error {
	limit := c.QueryInt("limit")
	if limit < 10 || limit > 100 {
		limit = 10
	}

	query := c.Query("query")
	tags := strings.Split(c.Query("tags"), ",")

	var requestCursor PostsCursor
	if err := decodeBase64AndUnmarshalJson(&requestCursor, c.Query("cursor")); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid cursor format")
	}

	repoPosts, err := queries.GetPosts(context.Background(), repository.GetPostsParams{
		CreatedAt: requestCursor.CreatedAt,
		Query:     query,
		Tags:      tags,
		Limit:     int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting user posts: %v", err)
	}

	var encodedResponseCursor string
	hasMore := limit < len(repoPosts)
	if hasMore {
		responseCursor := PostsCursor{
			CreatedAt: repoPosts[limit].CreatedAt,
		}
		encodedResponseCursor, err = marshalJsonAndEncodeBase64(responseCursor)
		if err != nil {
			return fmt.Errorf("error encoding cursor: %v", err)
		}
		repoPosts = repoPosts[:limit]
	}

	posts := make([]PostPayload, 0, len(repoPosts))
	for _, repoPost := range repoPosts {
		posts = append(posts, PostPayload{
			ID:        repoPost.ID,
			UserID:    repoPost.UserID,
			Title:     repoPost.Title,
			Content:   repoPost.Content,
			CreatedAt: repoPost.CreatedAt,
			Answered:  repoPost.Answered,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"posts":      posts,
		"cursor":     encodedResponseCursor,
		"hasMore":    hasMore,
		"totalCount": len(posts),
	})
}
