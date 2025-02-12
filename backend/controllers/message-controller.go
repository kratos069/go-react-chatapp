package controllers

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/chat-app/database"
	"github.com/chat-app/models"
	"github.com/chat-app/utils"
	"github.com/gofiber/fiber/v2"
)

// GetUsersForSidebar retrieves all users except the logged-in user
func GetUsersForSidebar(c *fiber.Ctx) error {
	// Get logged-in user ID
	claims, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID := claims.ID

	// Fetch users excluding the logged-in user
	var users []models.User
	if err := database.DB.Where("id != ?", userID).Select(
		"id, email, full_name, profile_pic, created_at, updated_at").
		Find(&users).Error; err != nil {
		log.Println("Error fetching users for sidebar:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Send the users as JSON response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": users,
	})
}

func GetMessages(c *fiber.Ctx) error {
	userToChatIDParam := c.Params("id")
	if userToChatIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	userToChatID, err := strconv.Atoi(userToChatIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid User ID",
		})
	}

	claims, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	userID := claims.ID

	var messages []models.Message
	err = database.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, userToChatID, userToChatID, userID,
	).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		log.Println("Error fetching messages:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messages": messages,
	})
}


// SendMessage handles sending a message (including text and image upload)
func SendMessage(c *fiber.Ctx) error {
	var req struct {
		Text  string `json:"text"`
		Image string `json:"image"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request data",
		})
	}

	claims, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	userID := claims.ID

	receiverID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid receiver ID",
		})
	}

	if req.Text == "" && req.Image == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Message text or image is required",
		})
	}

	var imageUrl string
	if req.Image != "" {
		cloudinaryService, err := utils.NewCloudinaryService()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to initialize Cloudinary service",
			})
		}

		imageUrl, err = cloudinaryService.UploadImage(context.Background(), req.Image)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload image",
			})
		}
	}

	message := models.Message{
		SenderID:   userID,
		ReceiverID: uint(receiverID),
		Text:       req.Text,
		Image:      imageUrl,
		CreatedAt:  time.Now(),
	}

	if err := database.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save message",
		})
	}

// Notify receiver via WebSocket
receiverSocket := utils.GetReceiverSocket(receiverID)
if receiverSocket != nil {
	err := receiverSocket.WriteJSON(fiber.Map{
		"event":   "newMessage",
		"message": message,
	})
	if err != nil {
		log.Printf("Error sending WebSocket message to user %d: %v\n", receiverID, err)
		utils.RemoveReceiverSocket(receiverID) // Clean up stale connection
	}
}


	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": message,
	})
}

