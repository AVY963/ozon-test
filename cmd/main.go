package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"ozon-posts/internal/repositories/inmemory"
	"ozon-posts/internal/repositories/postgres"
	"ozon-posts/internal/services"
	"ozon-posts/pkg/logger"
	"syscall"
	"time"

	"ozon-posts/internal/config"
	"ozon-posts/internal/handlers/graphql"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Load()

	l := logger.NewLogger(cfg)

	l.WithFields(logrus.Fields{
		"config": cfg,
	}).Info("Запуск приложения")

	var (
		userRepo    services.UserRepository
		postRepo    services.PostRepository
		commentRepo services.CommentRepository
		db          *sqlx.DB
	)

	if cfg.Database.IsPostgresMode() {
		l.Info("Инициализация PostgreSQL репозиториев")

		var err error
		db, err = postgres.InitPostgres(cfg, l)
		if err != nil {
			l.WithError(err).Fatal("Ошибка подключения к PostgreSQL")
		}
		defer db.Close()

		userRepo = postgres.NewUserRepository(db, l)
		postRepo = postgres.NewPostRepository(db, l)
		commentRepo = postgres.NewCommentRepository(db, l)
	} else {
		l.Info("Инициализация in-memory репозиториев")

		userRepo = inmemory.NewUserRepository(l)
		postRepo = inmemory.NewPostRepository(l)
		commentRepo = inmemory.NewCommentRepository(l)

		l.Info("In-memory репозитории успешно инициализированы")
	}

	userService := services.NewUserService(userRepo, l)
	postService := services.NewPostService(postRepo, userRepo, l)
	commentService := services.NewCommentService(
		commentRepo,
		postRepo,
		userRepo,
		l,
	)

	srv := graphql.InitGraphQLServer(userService, postService, commentService, l)

	mux := http.NewServeMux()

	mux.Handle("/query", srv)

	httpServer := &http.Server{
		Addr:    cfg.GetServerAddr(),
		Handler: mux,
	}

	go func() {
		l.WithField("addr", cfg.GetServerAddr()).Info("Сервер запущен")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.WithError(err).Fatal("Ошибка запуска HTTP сервера")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		l.WithError(err).Error("Принудительное завершение сервера")
	}

	l.Info("Сервер остановлен")
}
