package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkaAdapter "loyalty-service/internal/adapters/kafka"
	httpAdapter "loyalty-service/internal/adapters/http"
	"loyalty-service/internal/adapters/repository"
	"loyalty-service/internal/config"
	"loyalty-service/internal/core/domain"
	"loyalty-service/internal/core/services"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// Connect to database — retry until MySQL is ready
	var db *gorm.DB
	for i := range 10 {
		var err error
		db, err = gorm.Open(mysql.Open(cfg.DB.DSN()), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("db not ready (attempt %d/10): %v — retrying in 3s", i+1, err)
		time.Sleep(3 * time.Second)
		if i == 9 {
			log.Fatalf("failed to connect to database after 10 attempts: %v", err)
		}
	}

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

	// Start Kafka consumer + producer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	producer := kafkaAdapter.NewProducer(kafkaAdapter.ProducerConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	defer producer.Close()

	consumer := kafkaAdapter.NewConsumer(kafkaAdapter.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
		GroupID: cfg.Kafka.GroupID,
	}, orderSvc)
	consumer.Start(ctx)

	// Graceful shutdown on SIGINT / SIGTERM
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("shutting down...")
		cancel()
	}()

	// Start HTTP server
	app := httpAdapter.NewRouter(pointSvc, producer)
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("loyalty-service starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
