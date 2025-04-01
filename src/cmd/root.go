package cmd

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest/helpers"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/websocket"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/services"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	EmbedIndex embed.FS
	EmbedViews embed.FS
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Send free whatsapp API",
	Long: `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice, 
you can send whatsapp over http api but your whatsapp account have to be multi device version`,
	Run: runRest,
}

func init() {
	// Load environment variables first
	utils.LoadConfig(".")

	// Initialize configurations, flag is higher priority than env
	initEnvConfig()
	initFlags()
}

// initEnvConfig loads configuration from environment variables
func initEnvConfig() {
	// Application settings
	if envPort := viper.GetString("APP_PORT"); envPort != "" {
		config.AppPort = envPort
	}
	if envDebug := viper.GetBool("APP_DEBUG"); envDebug {
		config.AppDebug = envDebug
	}
	if envOs := viper.GetString("APP_OS"); envOs != "" {
		config.AppOs = envOs
	}
	if envBasicAuth := viper.GetString("APP_BASIC_AUTH"); envBasicAuth != "" {
		credential := strings.Split(envBasicAuth, ",")
		config.AppBasicAuthCredential = credential
	}
	if envChatFlushInterval := viper.GetInt("APP_CHAT_FLUSH_INTERVAL"); envChatFlushInterval > 0 {
		config.AppChatFlushIntervalDays = envChatFlushInterval
	}

	// Database settings
	if envDBURI := viper.GetString("DB_URI"); envDBURI != "" {
		config.DBURI = envDBURI
	}

	// WhatsApp settings
	if envAutoReply := viper.GetString("WHATSAPP_AUTO_REPLY"); envAutoReply != "" {
		config.WhatsappAutoReplyMessage = envAutoReply
	}
	if envWebhook := viper.GetString("WHATSAPP_WEBHOOK"); envWebhook != "" {
		webhook := strings.Split(envWebhook, ",")
		config.WhatsappWebhook = webhook
	}
	if envWebhookSecret := viper.GetString("WHATSAPP_WEBHOOK_SECRET"); envWebhookSecret != "" {
		config.WhatsappWebhookSecret = envWebhookSecret
	}
	if envAccountValidation := viper.GetBool("WHATSAPP_ACCOUNT_VALIDATION"); envAccountValidation {
		config.WhatsappAccountValidation = envAccountValidation
	}
	if envChatStorage := viper.GetBool("WHATSAPP_CHAT_STORAGE"); !envChatStorage {
		config.WhatsappChatStorage = envChatStorage
	}

	// Admin settings
	if envAdminUsername := viper.GetString("ADMIN_USERNAME"); envAdminUsername != "" {
		config.AdminUsername = envAdminUsername
	}
	if envAdminPassword := viper.GetString("ADMIN_PASSWORD"); envAdminPassword != "" {
		config.AdminPassword = envAdminPassword
	}
}

// initFlags sets up command line flags that override environment variables
func initFlags() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Application flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AppPort,
		"port", "p",
		config.AppPort,
		"change port number with --port <number> | example: --port=8080",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.AppDebug,
		"debug", "d",
		config.AppDebug,
		"hide or displaying log with --debug <true/false> | example: --debug=true",
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AppOs,
		"os", "",
		config.AppOs,
		`os name --os <string> | example: --os="Chrome"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.AppBasicAuthCredential,
		"basic-auth", "b",
		config.AppBasicAuthCredential,
		"basic auth credential | -b=yourUsername:yourPassword",
	)
	rootCmd.PersistentFlags().IntVarP(
		&config.AppChatFlushIntervalDays,
		"chat-flush-interval", "",
		config.AppChatFlushIntervalDays,
		`the interval to flush the chat storage --chat-flush-interval <number> | example: --chat-flush-interval=7`,
	)

	// Database flags
	rootCmd.PersistentFlags().StringVarP(
		&config.DBURI,
		"db-uri", "",
		config.DBURI,
		`the database uri to store the connection data database uri (by default, we'll use sqlite3 under storages/whatsapp.db). database uri --db-uri <string> | example: --db-uri="file:storages/whatsapp.db?_foreign_keys=off or postgres://user:password@localhost:5432/whatsapp"`,
	)

	// WhatsApp flags
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappAutoReplyMessage,
		"autoreply", "",
		config.WhatsappAutoReplyMessage,
		`auto reply when received message --autoreply <string> | example: --autoreply="Don't reply this message"`,
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&config.WhatsappWebhook,
		"webhook", "w",
		config.WhatsappWebhook,
		`forward event to webhook --webhook <string> | example: --webhook="https://yourcallback.com/callback"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.WhatsappWebhookSecret,
		"webhook-secret", "",
		config.WhatsappWebhookSecret,
		`secure webhook request --webhook-secret <string> | example: --webhook-secret="super-secret-key"`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappAccountValidation,
		"account-validation", "",
		config.WhatsappAccountValidation,
		`enable or disable account validation --account-validation <true/false> | example: --account-validation=true`,
	)
	rootCmd.PersistentFlags().BoolVarP(
		&config.WhatsappChatStorage,
		"chat-storage", "",
		config.WhatsappChatStorage,
		`enable or disable chat storage --chat-storage <true/false>. If you disable this, reply feature maybe not working properly | example: --chat-storage=true`,
	)

	// Admin flags
	rootCmd.PersistentFlags().StringVarP(
		&config.AdminUsername,
		"admin-username", "",
		config.AdminUsername,
		`admin username for master dashboard --admin-username <string> | example: --admin-username="admin"`,
	)
	rootCmd.PersistentFlags().StringVarP(
		&config.AdminPassword,
		"admin-password", "",
		config.AdminPassword,
		`admin password for master dashboard --admin-password <string> | example: --admin-password="secret"`,
	)
}

func runRest(_ *cobra.Command, _ []string) {
	if config.AppDebug {
		config.WhatsappLogLevel = "DEBUG"
	}

	// TODO: Init Rest App
	//preparing folder if not exist
	err := utils.CreateFolder(config.PathQrCode, config.PathSendItems, config.PathStorages, config.PathMedia)
	if err != nil {
		log.Fatalln(err)
	}

	engine := html.NewFileSystem(http.FS(EmbedIndex), ".html")
	engine.AddFunc("isEnableBasicAuth", func(token any) bool {
		return token != nil
	})
	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: int(config.WhatsappSettingMaxVideoSize),
	})

	app.Static("/statics", "./statics")
	app.Use("/components", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/components",
		Browse:     true,
	}))
	app.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/assets",
		Browse:     true,
	}))

	app.Use(middleware.Recovery())
	app.Use(middleware.BasicAuth())
	if config.AppDebug {
		app.Use(logger.New())
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Initialize database connection
	db, err := sql.Open(getDBDriver(), config.DBURI)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalln("Failed to run migrations:", err)
	}

	// Initialize WhatsApp database and client
	dbContainer := whatsapp.InitWaDB()
	cli := whatsapp.InitWaCLI(dbContainer)
	whatsappManager := services.GetWhatsAppManager()

	// Service
	appService := services.NewAppService(cli, dbContainer)
	sendService := services.NewSendService(cli, appService)
	userService := services.NewUserService(cli)
	messageService := services.NewMessageService(cli)
	groupService := services.NewGroupService(cli)
	newsletterService := services.NewNewsletterService(cli)
	adminService := services.NewAdminService(db)

	// Set admin service in middleware
	middleware.SetAdminService(adminService)

	// Initialize all active connections
	if err := initializeConnections(db, whatsappManager); err != nil {
		log.Printf("Warning: Failed to initialize some connections: %v", err)
	}

	// Rest
	rest.InitRestApp(app, appService)
	rest.InitRestSend(app, sendService)
	rest.InitRestUser(app, userService)
	rest.InitRestMessage(app, messageService)
	rest.InitRestGroup(app, groupService)
	rest.InitRestNewsletter(app, newsletterService)
	rest.InitRestAdmin(app, adminService)

	// Admin routes with authentication
	admin := app.Group("/admin")
	admin.Use(middleware.AdminAuth())
	admin.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/admin/index", fiber.Map{
			"AppHost":        fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname()),
			"AppVersion":     config.AppVersion,
			"BasicAuthToken": c.UserContext().Value(middleware.AuthorizationValue("BASIC_AUTH")),
		})
	})

	// Admin login page (no auth required)
	app.Get("/admin/login", func(c *fiber.Ctx) error {
		return c.Render("views/admin/login", fiber.Map{})
	})

	// Company dashboard route
	app.Get("/:companyId", func(c *fiber.Ctx) error {
		companyId := c.Params("companyId")
		token := c.UserContext().Value(middleware.AuthorizationValue("BASIC_AUTH")).(string)

		// Verify company exists and user has access
		if !adminService.VerifyCompanyAccess(companyId, token) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		return c.Render("views/index", fiber.Map{
			"AppHost":        fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname()),
			"AppVersion":     config.AppVersion,
			"BasicAuthToken": token,
			"MaxFileSize":    humanize.Bytes(uint64(config.WhatsappSettingMaxFileSize)),
			"MaxVideoSize":   humanize.Bytes(uint64(config.WhatsappSettingMaxVideoSize)),
			"CompanyId":      companyId,
		})
	})

	websocket.RegisterRoutes(app, appService)
	go websocket.RunHub()

	// Set auto reconnect to whatsapp server after booting
	go helpers.SetAutoConnectAfterBooting(appService)
	// Set auto reconnect checking
	go helpers.SetAutoReconnectChecking(cli)
	// Start auto flush chat csv
	if config.WhatsappChatStorage {
		go helpers.StartAutoFlushChatStorage()
	}

	if err = app.Listen(":" + config.AppPort); err != nil {
		log.Fatalln("Failed to start: ", err.Error())
	}
}

// getDBDriver returns the appropriate database driver based on the URI
func getDBDriver() string {
	if strings.HasPrefix(config.DBURI, "file:") {
		return "sqlite3"
	} else if strings.HasPrefix(config.DBURI, "postgres:") {
		return "postgres"
	}
	return "sqlite3" // default to sqlite3
}

// runMigrations executes database migrations
func runMigrations(db *sql.DB) error {
	// Read and execute migration files
	migrationFiles := []string{
		"src/database/migrations/001_create_companies_and_connections.sql",
	}

	for _, file := range migrationFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", file, err)
		}
	}

	return nil
}

// initializeConnections initializes all active WhatsApp connections
func initializeConnections(db *sql.DB, manager *services.WhatsAppManager) error {
	rows, err := db.Query(`
		SELECT c.user_id, c.token, comp.api_key 
		FROM connections c 
		JOIN companies comp ON c.company_id = comp.id 
		WHERE c.status = 'connected'
	`)
	if err != nil {
		return fmt.Errorf("failed to query active connections: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, token, apiKey string
		if err := rows.Scan(&userID, &token, &apiKey); err != nil {
			log.Printf("Error scanning connection row: %v", err)
			continue
		}

		// Convert userID to UUID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			log.Printf("Error parsing userID as UUID: %v", err)
			continue
		}

		// Get or create WhatsApp client instance
		if _, exists := manager.GetInstance(userUUID); !exists {
			// Here you would need to create a new client instance
			// This depends on your WhatsApp client initialization logic
			continue
		}

		// Update connection status
		if _, err := db.Exec("UPDATE connections SET status = 'connected' WHERE user_id = $1", userID); err != nil {
			log.Printf("Failed to update connection status for user %s: %v", userID, err)
		}
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(embedIndex embed.FS, embedViews embed.FS) {
	EmbedIndex = embedIndex
	EmbedViews = embedViews
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
