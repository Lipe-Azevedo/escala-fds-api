package database

import (
	"escala-fds-api/internal/entity"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewMySQLConnection() (*gorm.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		user, password, host, port, dbname)

	logLevel := gormlogger.Silent
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = gormlogger.Info
	}

	newLogger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(
		&entity.User{},
		&entity.Swap{},
		&entity.Comment{},
		&entity.Holiday{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return db, nil
}
