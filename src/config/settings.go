package config

import (
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
)

var (
	AppPlatform                          = waCompanionReg.DeviceProps_PlatformType(1)
	PathChatStorage                      = "storages/chat.csv"
	WhatsappSettingMaxImageSize    int64 = 20000000  // 20MB
	WhatsappSettingMaxDownloadSize int64 = 500000000 // 500MB
	WhatsappTypeUser                     = "@s.whatsapp.net"
	WhatsappTypeGroup                    = "@g.us"

	// MinIO Configuration
	MinIOEndpoint   = "localhost:9000"
	MinIOAccessKey  = "minioadmin"
	MinIOSecretKey  = "minioadmin"
	MinIOUseSSL     = false
	MinIOBucketName = "whatsapp-media"

	// RabbitMQ Configuration
	RabbitMQHost     = "localhost"
	RabbitMQPort     = "5672"
	RabbitMQUsername = "guest"
	RabbitMQPassword = "guest"
	RabbitMQVHost    = "/"

	// JWT Configuration
	JWTSecret     = "your-jwt-secret"
	JWTExpiration = "24h"
)
