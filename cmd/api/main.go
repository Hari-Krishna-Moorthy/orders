package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"

	appcfg "github.com/Hari-Krishna-Moorthy/orders/internals/platform/config"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/controller"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/middleware"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/pubsub"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/repository"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/service"
)

func main() {
	cfg := appcfg.New()

	// Singletons
	repo := repository.NewUserRepository()
	authSvc := service.NewAuthService(repo)
	psMgr := pubsub.NewManager()
	authCtrl := controller.NewAuthController(authSvc)
	wsCtrl := controller.NewWSController(psMgr)
	topicCtrl := controller.NewTopicController(psMgr)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Use(fiberlogger.New())

	// Health
	start := time.Now()
	app.Get("/health", func(c *fiber.Ctx) error {
		uptime := int(time.Since(start).Seconds())
		topics, subs := psMgr.Counts()
		return c.JSON(fiber.Map{
			"uptime_sec":  uptime,
			"topics":      topics,
			"subscribers": subs,
		})
	})

	// Auth
	auth := app.Group("/auth")
	auth.Post("/signup", authCtrl.SignUp)
	auth.Post("/signin", authCtrl.SignIn)
	auth.Post("/signout", middleware.JWT(), authCtrl.SignOut)

	// Topics (REST)
	app.Post("/topics", topicCtrl.Create)
	app.Delete("/topics/:name", topicCtrl.Delete)
	app.Get("/topics", topicCtrl.List)
	app.Get("/stats", topicCtrl.Stats)

	// WebSocket
	wsCtrl.Register(app)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	go func() {
		if err := app.Listen(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
}
