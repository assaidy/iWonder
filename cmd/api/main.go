package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	h "github.com/assaidy/iWonder/internals/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/joho/godotenv/autoload"
)

// TODO: use tx in all delete handlers if we check before deleting

func errorHandler(c *fiber.Ctx, err error) error {
	// code := fiber.StatusInternalServerError
	// var fiberErr *fiber.Error
	// if errors.As(err, &fiberErr) {
	// 	code = fiberErr.Code
	// }
	// c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	// NOTE: Logging occurs before this error handler is executed, so the internal error
	// has already been logged. We avoid exposing internal error details to the client
	// by returning a generic error message.
	// if code == fiber.StatusInternalServerError {
	// 	return c.SendStatus(code)
	// }
	// return c.Status(code).SendString(err.Error())

	return c.SendStatus(fiber.StatusInternalServerError)
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "iWonder",
		ServerHeader: "iWonder",
		Prefork:      false,
		ErrorHandler: errorHandler,
	})

	v1 := app.Group("/v1", logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error} | ${respHeader:Content-Type} | ${resBody}\n",
	}))
	{
		v1.Post("/users/register", h.HandleRegister)
		v1.Post("/users/login", h.HandleLogin)
		v1.Post("/users/access_token", h.HandleGetAccessToken)
		v1.Get("/users/id/:user_id", h.HandleGetUserByID)
		v1.Get("/users/username/:username", h.HandleGetUserByUsername)
		v1.Put("/users", h.WithJwt, h.HandleUpdateUser)
		v1.Delete("/users", h.WithJwt, h.HandleDeleteUser)

		v1.Post("/posts", h.WithJwt, h.HandleCreatePost)
		v1.Get("/posts/:post_id", h.HandleGetPost)
		v1.Put("/posts/:post_id", h.WithJwt, h.HandleUpdatePost)
		v1.Delete("/posts/:post_id", h.WithJwt, h.HandleDeletePost)

		// TEST:
		v1.Post("/posts/:post_id/tags", h.WithJwt, h.HandleAddPostTags)
		v1.Get("/posts/:post_id/tags", h.HandleGetPostTags)
		v1.Delete("/posts/:post_id/tags/:tag_name", h.WithJwt, h.HandleDeletePostTag)

		v1.Post("/posts/:post_id/comments", h.WithJwt, h.HandleCreateComment)
		v1.Put("/posts/comments/:comment_id", h.WithJwt, h.HandleUpdateComment)
		v1.Delete("/posts/comments/:comment_id", h.WithJwt, h.HandleDeleteComment)
		v1.Get("/posts/:post_id/comments", h.HandleGetAllPostComments)

		v1.Post("/posts/comments/:comment_id/votes", h.WithJwt, h.HandleVoteComment)
		v1.Delete("/posts/comments/:comment_id/votes", h.WithJwt, h.HandleUnvoteComment)
		v1.Get("/posts/comments/:comment_id/votes", h.HandleGetCommentVoteCounts)
		v1.Post("/posts/:post_id/answer", h.WithJwt, h.HandleSetPostAnswer)
		v1.Delete("/posts/:post_id/answer", h.WithJwt, h.HandleUnsetPostAnswer)
		v1.Get("/posts/:post_id/answer", h.HandleGetPostAnswer)

		v1.Get("users/:user_id/posts", h.HandleGetAllPostsForUser)
		v1.Get("/posts", h.HandleGetAllPosts) // ?query=xyz&tags=x,y,z
	}

	go func() {
		if err := app.Listen(os.Getenv("SERVER_ADDR")); err != nil {
			log.Fatal("error starting server: ", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		slog.Error("error shutting down server", "err", err, "pid", os.Getpid())
	} else {
		slog.Info("server shutdown completed gracefully", "pid", os.Getpid())
	}
}
