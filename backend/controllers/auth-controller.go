package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chat-app/database"
	"github.com/chat-app/models"
	"github.com/chat-app/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func SignupHandler(c *fiber.Ctx) error {
	type request struct {
		FullName string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var user request

	// Parse the body input to &user
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	// Check if fields are empty
	if user.FullName == "" || user.Email == "" ||
		user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	// Validate email format
	if !strings.Contains(user.Email, "@") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is incorrect",
		})
	}

	// Validate password length
	if len(user.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password should be at least 6 characters",
		})
	}

	// Check if user email already exists in the database
	var existingUser models.User
	if err := database.DB.Where("email = ?", user.Email).First(
		&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email already in use",
		})
	}

	// Hashing the Password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password),
		bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not save the password",
		})
	}

	// Create the new user
	newUser := models.User{
		FullName: user.FullName,
		Email:    user.Email,
		Password: string(passwordHash),
	}

	// Save the user to the DB
	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not create user",
		})
	}

	// Generate the JWT Token
	token, err := utils.CreateJWT(c, newUser.ID, newUser.FullName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	// Return success response with token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": fiber.Map{
			"id":         newUser.ID,
			"fullname":   newUser.FullName,
			"email":      newUser.Email,
			"created_at": newUser.CreatedAt,
			"updated_at": newUser.UpdatedAt,
		},
		"token": token,
	})
}

func LogoutHandler(c *fiber.Ctx) error {
	// Validate cookie name
	if utils.CookieName == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "server misconfiguration: cookie name is not set",
		})
	}

	// Determine cookie security based on environment
	isProduction := os.Getenv("ENV_KEY") == "production"

	// Clear the auth_token cookie by setting its expiry in the past
	c.Cookie(&fiber.Cookie{
		Name:     utils.CookieName,                // Name of the authentication cookie
		Value:    "",                              // Empty value to clear the cookie
		Expires:  time.Now().Add(-24 * time.Hour), // Set expiration in the past to clear the cookie
		HTTPOnly: true,                            // Ensure the cookie cannot be accessed via JavaScript
		Secure:   isProduction,                    // Set to `true` if running on HTTPS
		SameSite: fiber.CookieSameSiteStrictMode,  // SameSite policy for added security
	})

	// Return a success response
	if err := c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "successfully logged out",
	}); err != nil {
		// Log the error and return a response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to log out",
		})
	}

	return nil
}

func LoginHandler(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var user request

	// Parsing the input
	err := c.BodyParser(&user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unable to parse the input",
		})
	}

	// Check if fields are empty
	if user.Email == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password are required",
		})
	}

	// Fetch the user by email from the database
	var existingUser models.User
	if err := database.DB.Where("email = ?", user.Email).First(
		&existingUser).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	}

	// Check Password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password),
		[]byte(user.Password))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "please enter the correct password",
		})
	}

	// Generate the JWT Token
	token, err := utils.CreateJWT(c, existingUser.ID, existingUser.FullName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not create JWT token",
		})
	}

	// Respond with user data and the JWT token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": fiber.Map{
			"id":         existingUser.ID,
			"fullname":   existingUser.FullName,
			"email":      existingUser.Email,
			"created_at": existingUser.CreatedAt,
			"updated_at": existingUser.UpdatedAt,
		},
		"token": token, // Send the token in the response as well (optional)
	})
}

func SignedInUser(c *fiber.Ctx) error {
	// Retrieve user from context and safely assert type
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not authenticated",
		})
	}

	// Retrieve the JWT token from the cookie
	token := c.Cookies("auth_token")

	// If token is not found, return an error
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token not found in cookies",
		})
	}

	// Return user information (extend if needed)
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":         user.ID,
			"fullname":   user.FullName,
			"email":      user.Email,
			"profilePic": user.ProfilePic,
		},
		"token": token,
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	// Extract user ID from request context
	claims, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	userID := claims.ID

	var user models.User

	// Find user in the database
	if err := database.DB.First(&user, userID).Error; err != nil {
		log.Println("Error finding user in database:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Initialize Cloudinary service
	cloudService, err := utils.NewCloudinaryService()
	if err != nil {
		log.Println("Error initializing Cloudinary service:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to cloud service",
		})
	}

	ctx := context.Background()

	// Handle profile picture upload
	var uploadedURL string
	if profilePic, err := c.FormFile("profilePic"); err == nil {
		// Delete the old profile picture if it exists
		if user.ProfilePic != "" {
			publicID := extractPublicID(user.ProfilePic)
			if err := cloudService.DeleteImage(ctx, publicID); err != nil {
				log.Println("Warning: Failed to delete old profile image:", err)
			}
		}

		// Save the uploaded file locally
		filePath := fmt.Sprintf("./uploads/%s", profilePic.Filename)
		if err := c.SaveFile(profilePic, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save the uploaded file",
			})
		}

		// Upload the image to Cloudinary
		uploadedURL, err = cloudService.UploadImage(ctx, filePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload profile image",
			})
		}
	}

	// Update the user's profile picture in the database
	if err := database.DB.Model(&user).Updates(models.User{
		ProfilePic: uploadedURL,
	}).Error; err != nil {
		log.Println("Error updating user profile:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// Respond with updated user data
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// Helper function to extract public ID from a Cloudinary URL
func extractPublicID(url string) string {
	// Assuming URL format: https://res.cloudinary.com/<cloud-name>/image/upload/v1234567890/folder/publicID.extension
	// Step 1: Split the URL into parts using "/" as the separator
	parts := strings.Split(url, "/")
	// Step 2: Get the last part of the URL, which contains
	// "publicID.extension"
	lastPart := parts[len(parts)-1]
	// Step 3: Split "my-image-name.extension" using "." to separate
	// the public ID from the file extension
	publicID := strings.Split(lastPart, ".")[0]
	return publicID
}
