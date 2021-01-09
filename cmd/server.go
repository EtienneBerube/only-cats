package cmd

import (
	"context"
	"github.com/EtienneBerube/only-cats/internal/handlers"
	"github.com/EtienneBerube/only-cats/internal/middleware"
	"github.com/EtienneBerube/only-cats/internal/repositories"
	"github.com/EtienneBerube/only-cats/pkg/config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func RunServer(config config.Config) {
	address := ":" + config.Port
	repositories.InitDB(config)

	gin.ForceConsoleColor()
	router := gin.New()

	router.Use(gin.Recovery())

	// Custom Logger
	router.Use(gin.LoggerWithFormatter(middleware.WithLogging))

	initRoutes(router)

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	log.Printf("Server is ready to handle requests at %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", address, err)
	}

	<-done
	log.Println("Server stopped")
}

func initRoutes(router *gin.Engine) {

	router.GET("/ping", handlers.Ping)

	router.POST("/login", handlers.Login)
	router.POST("/signup", handlers.SignUp)

	router.GET("/charger", handlers.GetChargers)
	router.GET("/charger/:id", handlers.GetChargerByID)

	router.GET("/user/:id/photo", handlers.GetProfilePicForUser)

	authenticated := router.Group("/") // Change authService from nil to smtg else
	{
		authenticated.Use(middleware.Auth())
		authenticated.GET("/user", handlers.GetCurrentUser)
		authenticated.GET("/user/:id", handlers.GetUserByID)
		authenticated.PUT("/user/:id", handlers.UpdateUser) // MIGHT REMOVE ID IF ONLY UPDATE TO CURRENT USER
		authenticated.POST("/user/:id/photo", handlers.GetProfilePicForUser)

		authenticated.POST("/charger", handlers.CreateCharger)

	}
}