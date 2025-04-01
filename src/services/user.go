package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"strconv"
	"time"

	domainUser "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/user"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database/models"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// UserService handles user-related operations
type UserService struct {
	db     *gorm.DB
	client *whatsmeow.Client
}

// NewUserServiceDB creates a new user service instance
func NewUserServiceDB(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// NewUserService creates a new user service instance
func NewUserService(client *whatsmeow.Client) *UserService {
	return &UserService{
		client: client,
	}
}

// List returns all users for a company
func (s *UserService) List(c *fiber.Ctx) error {
	companyID := c.Locals("companyID").(int64)
	users, err := models.GetActiveUsers(companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
	})
}

// Create creates a new user
func (s *UserService) Create(c *fiber.Ctx) error {
	companyID := c.Locals("companyID").(int64)

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user := &models.User{
		CompanyID: companyID,
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password,
		Active:    true,
	}

	if err := models.CreateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": user,
	})
}

// Get returns a user by ID
func (s *UserService) Get(c *fiber.Ctx) error {
	companyID := c.Locals("companyID").(int64)
	userID := c.Params("id")

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	user, err := models.GetUserByID(userIDInt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if user.CompanyID != companyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// Update updates a user
func (s *UserService) Update(c *fiber.Ctx) error {
	companyID := c.Locals("companyID").(int64)
	userID := c.Params("id")

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	user, err := models.GetUserByID(userIDInt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if user.CompanyID != companyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Active   bool   `json:"active"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user.Name = input.Name
	user.Email = input.Email
	if input.Password != "" {
		user.Password = input.Password
	}
	user.Active = input.Active

	if err := models.UpdateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// Delete deletes a user
func (s *UserService) Delete(c *fiber.Ctx) error {
	companyID := c.Locals("companyID").(int64)
	userID := c.Params("id")

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	user, err := models.GetUserByID(userIDInt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if user.CompanyID != companyID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	if err := models.DeleteUser(user.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// WhatsAppUserService handles WhatsApp user operations
type WhatsAppUserService struct {
	WaCli *whatsmeow.Client
}

// NewWhatsAppUserService creates a new WhatsAppUserService instance
func NewWhatsAppUserService(waCli *whatsmeow.Client) domainUser.IUserService {
	return &WhatsAppUserService{
		WaCli: waCli,
	}
}

func (service WhatsAppUserService) Info(ctx context.Context, request domainUser.InfoRequest) (response domainUser.InfoResponse, err error) {
	err = validations.ValidateUserInfo(ctx, request)
	if err != nil {
		return response, err
	}
	var jids []types.JID
	dataWaRecipient, err := whatsapp.ValidateJidWithLogin(service.WaCli, request.Phone)
	if err != nil {
		return response, err
	}

	jids = append(jids, dataWaRecipient)
	resp, err := service.WaCli.GetUserInfo(jids)
	if err != nil {
		return response, err
	}

	for _, userInfo := range resp {
		var device []domainUser.InfoResponseDataDevice
		for _, j := range userInfo.Devices {
			device = append(device, domainUser.InfoResponseDataDevice{
				User:   j.User,
				Agent:  j.RawAgent,
				Device: whatsapp.GetPlatformName(int(j.Device)),
				Server: j.Server,
				AD:     j.ADString(),
			})
		}

		data := domainUser.InfoResponseData{
			Status:    userInfo.Status,
			PictureID: userInfo.PictureID,
			Devices:   device,
		}
		if userInfo.VerifiedName != nil {
			data.VerifiedName = fmt.Sprintf("%v", *userInfo.VerifiedName)
		}
		response.Data = append(response.Data, data)
	}

	return response, nil
}

func (service WhatsAppUserService) Avatar(ctx context.Context, request domainUser.AvatarRequest) (response domainUser.AvatarResponse, err error) {
	chanResp := make(chan domainUser.AvatarResponse)
	chanErr := make(chan error)
	waktu := time.Now()

	go func() {
		err = validations.ValidateUserAvatar(ctx, request)
		if err != nil {
			chanErr <- err
		}
		dataWaRecipient, err := whatsapp.ValidateJidWithLogin(service.WaCli, request.Phone)
		if err != nil {
			chanErr <- err
		}
		pic, err := service.WaCli.GetProfilePictureInfo(dataWaRecipient, &whatsmeow.GetProfilePictureParams{
			Preview:     request.IsPreview,
			IsCommunity: request.IsCommunity,
		})
		if err != nil {
			chanErr <- err
		} else if pic == nil {
			chanErr <- errors.New("no avatar found")
		} else {
			response.URL = pic.URL
			response.ID = pic.ID
			response.Type = pic.Type

			chanResp <- response
		}
	}()

	for {
		select {
		case err := <-chanErr:
			return response, err
		case response := <-chanResp:
			return response, nil
		default:
			if waktu.Add(2 * time.Second).Before(time.Now()) {
				return response, pkgError.ContextError("Error timeout get avatar !")
			}
		}
	}
}

func (service WhatsAppUserService) MyListGroups(_ context.Context) (response domainUser.MyListGroupsResponse, err error) {
	whatsapp.MustLogin(service.WaCli)

	groups, err := service.WaCli.GetJoinedGroups()
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", groups)
	for _, group := range groups {
		response.Data = append(response.Data, *group)
	}
	return response, nil
}

func (service WhatsAppUserService) MyListNewsletter(_ context.Context) (response domainUser.MyListNewsletterResponse, err error) {
	whatsapp.MustLogin(service.WaCli)

	datas, err := service.WaCli.GetSubscribedNewsletters()
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", datas)
	for _, data := range datas {
		response.Data = append(response.Data, *data)
	}
	return response, nil
}

func (service WhatsAppUserService) MyPrivacySetting(_ context.Context) (response domainUser.MyPrivacySettingResponse, err error) {
	whatsapp.MustLogin(service.WaCli)

	resp, err := service.WaCli.TryFetchPrivacySettings(true)
	if err != nil {
		return
	}

	response.GroupAdd = string(resp.GroupAdd)
	response.Status = string(resp.Status)
	response.ReadReceipts = string(resp.ReadReceipts)
	response.Profile = string(resp.Profile)
	return response, nil
}

func (service WhatsAppUserService) MyListContacts(ctx context.Context) (response domainUser.MyListContactsResponse, err error) {
	whatsapp.MustLogin(service.WaCli)

	contacts, err := service.WaCli.Store.Contacts.GetAllContacts()
	if err != nil {
		return
	}

	for jid, contact := range contacts {
		response.Data = append(response.Data, domainUser.MyListContactsResponseData{
			JID:  jid,
			Name: contact.FullName,
		})
	}

	return response, nil
}

func (service WhatsAppUserService) ChangeAvatar(ctx context.Context, request domainUser.ChangeAvatarRequest) (err error) {
	whatsapp.MustLogin(service.WaCli)

	file, err := request.Avatar.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Read original image
	srcImage, err := imaging.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Get original dimensions
	bounds := srcImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions for 1:1 aspect ratio
	size := width
	if height < width {
		size = height
	}
	if size > 640 {
		size = 640
	}

	// Create a square crop from the center
	left := (width - size) / 2
	top := (height - size) / 2
	croppedImage := imaging.Crop(srcImage, image.Rect(left, top, left+size, top+size))

	// Resize if needed
	if size > 640 {
		croppedImage = imaging.Resize(croppedImage, 640, 640, imaging.Lanczos)
	}

	// Convert to bytes
	var buf bytes.Buffer
	err = imaging.Encode(&buf, croppedImage, imaging.JPEG, imaging.JPEGQuality(80))
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	_, err = service.WaCli.SetGroupPhoto(types.JID{}, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (service WhatsAppUserService) ChangePushName(ctx context.Context, request domainUser.ChangePushNameRequest) (err error) {
	whatsapp.MustLogin(service.WaCli)

	err = service.WaCli.SendAppState(appstate.BuildSettingPushName(request.PushName))
	if err != nil {
		return err
	}
	return nil
}
