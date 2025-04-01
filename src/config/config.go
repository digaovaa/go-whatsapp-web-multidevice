package config

var (
	// App settings
	AppPort                  = "8080"
	AppDebug                 = false
	AppOs                    = "Chrome"
	AppVersion               = "1.0.0"
	AppBasicAuthCredential   []string
	AppChatFlushIntervalDays = 7

	// Database settings
	DBURI = "file:storages/whatsapp.db?_foreign_keys=off"

	// WhatsApp settings
	WhatsappAutoReplyMessage    = ""
	WhatsappWebhook             []string
	WhatsappWebhookSecret              = ""
	WhatsappAccountValidation          = true
	WhatsappChatStorage                = true
	WhatsappLogLevel                   = "INFO"
	WhatsappSettingMaxFileSize  uint64 = 100 * 1024 * 1024 // 100MB
	WhatsappSettingMaxVideoSize uint64 = 16 * 1024 * 1024  // 16MB

	// Path settings
	PathQrCode    = "storages/qrcode"
	PathSendItems = "storages/senditems"
	PathStorages  = "storages"
	PathMedia     = "storages/media"

	// Admin settings
	AdminUsername = "admin"
	AdminPassword = "admin"
)
