package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const CookieName = "auth_token"

// CreateJWT generates a JWT token and sets it in a cookie
func CreateJWT(c *fiber.Ctx, userID uint, username string) (string, error) {
	// Load the secret key from the environment variable
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", fmt.Errorf("JWT_SECRET is not set in environment variables")
	}

	// Define token claims
	claims := jwt.MapClaims{
		"id":       userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":      time.Now().Unix(),                     // Issued at
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Printf("Error signing JWT token: %v", err)
		return "", err
	}

	// Determine cookie security based on environment
	isProduction := os.Getenv("ENV_KEY") == "production"

	// Set the token in a cookie
	c.Cookie(&fiber.Cookie{
		Name:     CookieName,  // Cookie name
		Value:    signedToken, // JWT token value
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,         // Prevent JavaScript access
		Secure:   isProduction, // Use secure cookies in production
		SameSite: "Strict",     // SameSite policy (prevent CSRF attacks)
	})

	return signedToken, nil
}

// ValidateToken validates a JWT token and extracts the claims
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Load the secret key from environment variables
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET is not set in environment variables")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	// Handle token parsing errors
	if err != nil {
		log.Printf("Error parsing JWT token: %v", err)
		return nil, err
	}

	// Validate the token claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Ensure the token is not expired
		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, errors.New("invalid expiration claim in token")
		}
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token is expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
