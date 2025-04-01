package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type MinIOSettings struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type RabbitMQSettings struct {
	Host     string
	Port     string
	Username string
	Password string
	VHost    string
}

var (
	minioConfig    MinIOSettings
	rabbitmqConfig RabbitMQSettings
)

func init() {
	initServicesConfig()
}

func initServicesConfig() {
	// MinIO Configuration
	minioConfig = MinIOSettings{
		Endpoint:        getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:          viper.GetBool("MINIO_USE_SSL"),
		BucketName:      getEnvOrDefault("MINIO_BUCKET_NAME", "whatsapp-media"),
	}

	// RabbitMQ Configuration
	rabbitmqConfig = RabbitMQSettings{
		Host:     getEnvOrDefault("RABBITMQ_HOST", "localhost"),
		Port:     getEnvOrDefault("RABBITMQ_PORT", "5672"),
		Username: getEnvOrDefault("RABBITMQ_USERNAME", "guest"),
		Password: getEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnvOrDefault("RABBITMQ_VHOST", "/"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetMinIOConnectionString returns the MinIO connection string
func GetMinIOConnectionString() string {
	protocol := "http"
	if minioConfig.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%s@%s", protocol, minioConfig.AccessKeyID, minioConfig.SecretAccessKey, minioConfig.Endpoint)
}

// GetRabbitMQConnectionString returns the RabbitMQ connection string
func GetRabbitMQConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s", rabbitmqConfig.Username, rabbitmqConfig.Password, rabbitmqConfig.Host, rabbitmqConfig.Port, rabbitmqConfig.VHost)
}
