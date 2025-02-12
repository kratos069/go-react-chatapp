package middleware

import (
	"os"

	"github.com/chat-app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm" // For database interaction
)

const CookieName = "auth_token"

// Secret key used for signing JWT
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// AuthMiddleware ensures the user is authenticated
func AuthMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the token from the cookie
		tokenStr := c.Cookies(CookieName)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Parse and validate the JWT token (got from cookie)
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Retrieve custom claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Get the user ID from claims
		userID, ok := claims["id"].(float64) // Use float64 because JWT encodes numbers as float64
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User ID not found in token",
			})
		}

		// Fetch user details from the database
		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve user information",
			})
		}

		// Attach the user object to the context for use in downstream handlers
		c.Locals("user", user)

		// Proceed to the next handler
		return c.Next()
	}
}
