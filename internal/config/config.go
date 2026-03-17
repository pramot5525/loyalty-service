package config

import (
	"fmt"
	"os"
	"strings"
)

func splitBrokers(s string) []string {
	return strings.Split(s, ",")
}

type Config struct {
	AppPort string
	DB      DBConfig
	Kafka   KafkaConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "3000"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "loyalty"),
			Password: getEnv("DB_PASSWORD", "loyalty_pass"),
			Name:     getEnv("DB_NAME", "loyalty_db"),
		},
		Kafka: KafkaConfig{
			Brokers: splitBrokers(getEnv("KAFKA_BROKERS", "localhost:9092")),
			Topic:   getEnv("KAFKA_TOPIC", "order.events"),
			GroupID: getEnv("KAFKA_GROUP_ID", "loyalty-service-group"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
