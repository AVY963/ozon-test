package postgres

import (
	"fmt"
	"time"

	"ozon-posts/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func InitPostgres(cfg *config.Config, logger *logrus.Logger) (*sqlx.DB, error) {
	dsn := cfg.Database.GetPostgresDSN()

	logger.WithFields(logrus.Fields{
		"host": cfg.Database.Postgres.Host,
		"port": cfg.Database.Postgres.Port,
		"user": cfg.Database.Postgres.User,
		"db":   cfg.Database.Postgres.DBName,
	}).Info("Подключение к PostgreSQL")

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к базе данных: %w", err)
	}

	logger.Info("Успешное подключение к PostgreSQL")
	return db, nil
}
