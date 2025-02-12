package main

import (
	"github.com/chat-app/controllers"
	"github.com/chat-app/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RoutesSetup(app *fiber.App, db *gorm.DB) {
	// Auth Routes
	app.Post("/api/auth/signup", controllers.SignupHandler)
	app.Post("/api/auth/logout", controllers.LogoutHandler)
	app.Post("/api/auth/login", controllers.LoginHandler)

	// AuthMiddleware ensures the user is authenticated (to proceed)
	app.Use(middleware.AuthMiddleware(db))
	// Now User will be available to be used in authenticated routes
	// and info can be passed through him
	app.Get("/api/auth/check", controllers.SignedInUser)

	// User Routes
	app.Put("/api/user/update-profile", controllers.UpdateProfile)

	// Message Routes
	app.Get("/api/messages/users", controllers.GetUsersForSidebar)
	app.Get("/api/messages/:id", controllers.GetMessages)
	app.Post("/api/messages/send/:id", controllers.SendMessage)
}
