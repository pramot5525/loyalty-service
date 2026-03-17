package main

import (
	"fmt"
	"log"

	"loyalty-service/internal/adapters/repository"
	"loyalty-service/internal/config"
	"loyalty-service/internal/core/domain"
	"loyalty-service/internal/core/services"

	httpAdapter "loyalty-service/internal/adapters/http"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load .env if present (ignore error when not found)
	_ = godotenv.Load()

	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(mysql.Open(cfg.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Order{},
		&domain.PointTransaction{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Wire repositories
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	pointTxRepo := repository.NewPointTransactionRepository(db)

	// Wire services
	orderSvc := services.NewOrderService(userRepo, orderRepo, pointTxRepo)
	pointSvc := services.NewPointService(userRepo)

	// Start HTTP server
	app := httpAdapter.NewRouter(orderSvc, pointSvc)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("loyalty-service starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
