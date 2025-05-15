package main

import (
	"errors"
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

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	// NOTE: Logging occurs before this error handler is executed, so the internal error
	// has already been logged. We avoid exposing internal error details to the client
	// by returning a generic error message.
	if code == fiber.StatusInternalServerError {
		return c.SendStatus(code)
	}
	return c.Status(code).SendString(err.Error())
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "iWonder",
		ServerHeader: "iWonder",
		Prefork:      false,
		ErrorHandler: errorHandler,
	})

	v1 := app.Group("/v1", logger.New())
	{
		v1.Post("/users/register", h.HandleRegister)
		v1.Post("/users/login", h.HandleLogin)
		v1.Post("/users/access_token", h.HandleGetAccessToken)
		v1.Get("/users/:user_id", h.HandleGetUserByID)
		v1.Get("/users/:username", h.HandleGetUserByUsername)
		v1.Put("/users", h.HandleUpdateUser, h.WithJwt)
		v1.Delete("/users", h.HandleDeleteUser, h.WithJwt)

		v1.Post("/posts", h.HandleCreatePost, h.WithJwt)
		v1.Get("/posts/:post_id", h.HandleGetPostByID)
		v1.Put("/posts/:post_id", h.HandleUpdatePostByID, h.WithJwt)
		v1.Patch("/posts/:post_id/answered", h.HandleTogglePostAnswered, h.WithJwt)
		v1.Delete("/posts/:post_id", h.HandleDeletePostByID, h.WithJwt)

		v1.Post("/posts/:post_id/tags", h.HandleAddPostTags, h.WithJwt)
		v1.Delete("/posts/:post_id/tags/:tag_name", h.HandleDeletePostTag, h.WithJwt)
		// comments (Upvote/downvote, mark as right answer)
		// search and filteration
		// report posts
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
