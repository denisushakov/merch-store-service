package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"merch-store-service/internal/api"
	coinServices "merch-store-service/internal/domain/coins/service"
	"merch-store-service/internal/domain/repository"
	userServices "merch-store-service/internal/domain/users/service"
	"merch-store-service/internal/infra/config"
	middleware "merch-store-service/internal/infra/http/middlewares"
	"merch-store-service/internal/infra/jwtutils"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Server *http.Server
	port   int
}

func New(cfg *config.Config) *App {
	storage, err := repository.New(cfg)
	if err != nil {
		log.Fatalf("failed to init storage %v", err)
	}

	jwtManager := jwtutils.NewJWTManager(cfg)

	userService := userServices.NewUserService(storage, jwtManager)
	coinService := coinServices.NewCoinService(storage)
	router := chi.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	router.Use(authMiddleware.Middleware())

	server := &Server{
		UserService: userService,
		CoinService: coinService,
	}

	apiHandler := api.HandlerFromMux(server, router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: apiHandler,
	}

	return &App{
		Server: srv,
		port:   cfg.Port,
	}

}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *App) Run() error {
	const op = "merch-store.Run"

	if err := a.Server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "merch-store.Stop"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("stopping merch-store port %d", a.port)
	if err := a.Server.Shutdown(ctx); err != nil {
		log.Fatalf("%s: Server shutdown failed: %v", op, err)
	}

	<-ctx.Done()
	log.Println("timeout of 5 seconds.")

	log.Println("Server gracefully stopped")
}
