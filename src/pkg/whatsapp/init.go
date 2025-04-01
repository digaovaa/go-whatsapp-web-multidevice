package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/websocket"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// Type definitions
type ExtractedMedia struct {
	MediaPath string `json:"media_path"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
}

type evtReaction struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type evtMessage struct {
	ID            string `json:"id,omitempty"`
	Text          string `json:"text,omitempty"`
	RepliedId     string `json:"replied_id,omitempty"`
	QuotedMessage string `json:"quoted_message,omitempty"`
}

// Global variables
var (
	log           waLog.Logger
	historySyncID int32
	startupTime   = time.Now().Unix()
)

// InitWaDB initializes the WhatsApp database connection
func InitWaDB() *sqlstore.Container {
	log = waLog.Stdout("Main", config.WhatsappLogLevel, true)
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)

	storeContainer, err := initDatabase(dbLog)
	if err != nil {
		log.Errorf("Database initialization error: %v", err)
		panic(pkgError.InternalServerError(fmt.Sprintf("Database initialization error: %v", err)))
	}

	return storeContainer
}

// initDatabase creates and returns a database store container based on the configured URI
func initDatabase(dbLog waLog.Logger) (*sqlstore.Container, error) {
	if strings.HasPrefix(config.DBURI, "file:") {
		return sqlstore.New("sqlite3", config.DBURI, dbLog)
	} else if strings.HasPrefix(config.DBURI, "postgres:") {
		return sqlstore.New("postgres", config.DBURI, dbLog)
	}

	return nil, fmt.Errorf("unknown database type: %s. Currently only sqlite3(file:) and postgres are supported", config.DBURI)
}

// InitWaCLI initializes a new WhatsApp client for a specific user
func InitWaCLI(storeContainer *sqlstore.Container, userID int64) (*whatsmeow.Client, error) {
	device, err := storeContainer.GetFirstDevice()
	if err != nil {
		log.Errorf("Failed to get device: %v", err)
		return nil, err
	}

	if device == nil {
		log.Errorf("No device found")
		return nil, fmt.Errorf("no device found")
	}

	// Configure device properties
	osName := fmt.Sprintf("%s %s", config.AppOs, config.AppVersion)
	store.DeviceProps.PlatformType = &config.AppPlatform
	store.DeviceProps.Os = &osName

	// Create and configure the client
	cli := whatsmeow.NewClient(device, waLog.Stdout("Client", config.WhatsappLogLevel, true))
	cli.EnableAutoReconnect = true
	cli.AutoTrustIdentity = true
	cli.AddEventHandler(func(rawEvt interface{}) {
		handler(rawEvt, userID)
	})

	// Add the instance to the manager
	GetInstanceManager().AddInstance(userID, cli)

	return cli, nil
}

// handler is the main event handler for WhatsApp events
func handler(rawEvt interface{}, userID int64) {
	switch evt := rawEvt.(type) {
	case *events.DeleteForMe:
		handleDeleteForMe(evt, userID)
	case *events.AppStateSyncComplete:
		handleAppStateSyncComplete(evt, userID)
	case *events.PairSuccess:
		handlePairSuccess(evt, userID)
	case *events.LoggedOut:
		handleLoggedOut(userID)
	case *events.Connected, *events.PushNameSetting:
		handleConnectionEvents(userID)
	case *events.StreamReplaced:
		handleStreamReplaced(userID)
	case *events.Message:
		handleMessage(evt, userID)
	case *events.Receipt:
		handleReceipt(evt, userID)
	case *events.Presence:
		handlePresence(evt, userID)
	case *events.HistorySync:
		handleHistorySync(evt, userID)
	case *events.AppState:
		handleAppState(evt, userID)
	}
}

// Event handler functions with userID parameter
func handleDeleteForMe(evt *events.DeleteForMe, userID int64) {
	log.Infof("Deleted message %s for %s", evt.MessageID, evt.SenderJID.String())
}

func handleAppStateSyncComplete(evt *events.AppStateSyncComplete, userID int64) {
	instance := GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return
	}

	if len(instance.Client.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
		if err := instance.Client.SendPresence(types.PresenceAvailable); err != nil {
			log.Errorf("Failed to send presence: %v", err)
		}
	}
}

func handlePairSuccess(evt *events.PairSuccess, userID int64) {
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "LOGIN_SUCCESS",
		Message: "Success login to whatsapp",
		Result:  nil,
	}
}

func handleLoggedOut(userID int64) {
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:   "LIST_DEVICES",
		Result: nil,
	}
	GetInstanceManager().RemoveInstance(userID)
}

func handleConnectionEvents(userID int64) {
	instance := GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return
	}

	if len(instance.Client.Store.PushName) == 0 {
		return
	}

	if err := instance.Client.SendPresence(types.PresenceAvailable); err != nil {
		log.Errorf("Failed to send presence: %v", err)
	}
}

func handleStreamReplaced(userID int64) {
	os.Exit(0)
}

func handleMessage(evt *events.Message, userID int64) {
	// Get instance for this user
	instance := GetInstanceManager().GetInstance(userID)
	if instance == nil {
		logrus.Errorf("No instance found for user %d", userID)
		return
	}

	// Build message metadata
	metaParts := buildMessageMetaParts(evt)
	logrus.Infof("Received message from %s: %s", evt.Info.SourceString(), metaParts)

	// Handle different message types
	if evt.Message.GetImageMessage() != nil {
		handleImageMessage(evt, userID)
	}

	// Forward to webhook if configured
	if len(config.WhatsappWebhook) > 0 {
		if err := forwardToWebhook(evt, userID); err != nil {
			logrus.Errorf("Failed to forward message to webhook: %v", err)
		}
	}
}

func handleReceipt(evt *events.Receipt, userID int64) {
	if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
		log.Infof("%v was read by %s at %s", evt.MessageIDs, evt.SourceString(), evt.Timestamp)
	}
}

func handlePresence(evt *events.Presence, userID int64) {
	if evt.Unavailable {
		if evt.LastSeen.IsZero() {
			log.Infof("%s is now offline", evt.From)
		} else {
			log.Infof("%s is now offline (last seen: %s)", evt.From, evt.LastSeen)
		}
	} else {
		log.Infof("%s is now online", evt.From)
	}
}

func handleHistorySync(evt *events.HistorySync, userID int64) {
	instance := GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return
	}

	id := atomic.AddInt32(&historySyncID, 1)
	fileName := fmt.Sprintf("%s/history-%d-%s-%d-%s.json",
		config.PathStorages,
		startupTime,
		evt.Data.SyncType,
		id,
		instance.Client.Store.ID.String(),
	)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Errorf("Failed to open file to write history sync: %v", err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(evt.Data)
	if err != nil {
		log.Errorf("Failed to write history sync: %v", err)
	}
	file.Close()
}

func handleAppState(evt *events.AppState, userID int64) {
	log.Debugf("App state event: %+v / %+v", evt.Index, evt.SyncActionValue)
}

func buildMessageMetaParts(evt *events.Message) string {
	var metaParts []string
	if evt.Info.IsFromMe {
		metaParts = append(metaParts, "from me")
	}
	if evt.Info.IsGroup {
		metaParts = append(metaParts, "from group")
	}
	if evt.Info.IsIncomingBroadcast() {
		metaParts = append(metaParts, "from broadcast")
	}
	if evt.Message.GetProtocolMessage() != nil && evt.Message.GetProtocolMessage().GetType() == waE2E.ProtocolMessage_EPHEMERAL_SETTING {
		metaParts = append(metaParts, "from status")
	}
	if evt.Message.GetEphemeralMessage() != nil {
		metaParts = append(metaParts, "ephemeral")
	}
	return strings.Join(metaParts, ", ")
}

func handleImageMessage(evt *events.Message, userID int64) {
	instance := GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return
	}

	if img := evt.Message.GetImageMessage(); img != nil {
		if path, err := ExtractMedia(config.PathStorages, img, instance.Client); err != nil {
			log.Errorf("Failed to download image: %v", err)
		} else {
			log.Infof("Image downloaded to %s", path)
		}
	}
}

func handleAutoReply(evt *events.Message, client *whatsmeow.Client) {
	if config.WhatsappAutoReplyMessage != "" &&
		!isGroupJid(evt.Info.Chat.String()) &&
		!evt.Info.IsIncomingBroadcast() &&
		evt.Message.GetExtendedTextMessage().GetText() != "" {
		_, _ = client.SendMessage(
			context.Background(),
			FormatJID(evt.Info.Sender.String()),
			&waE2E.Message{Conversation: proto.String(config.WhatsappAutoReplyMessage)},
		)
	}
}

func handleWebhookForward(evt *events.Message, userID int64) error {
	if len(config.WhatsappWebhook) > 0 &&
		!strings.Contains(evt.Info.SourceString(), "broadcast") &&
		!isFromMySelf(evt.Info.SourceString(), GetInstanceManager().GetInstance(userID).Client) {
		go func(evt *events.Message) {
			if err := forwardToWebhook(evt, userID); err != nil {
				logrus.Error("Failed forward to webhook: ", err)
			}
		}(evt)
	}
	return nil
}
