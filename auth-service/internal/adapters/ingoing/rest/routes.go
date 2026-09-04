package rest

import (
	"auth-service/internal/application/ports"

	"github.com/gin-gonic/gin"
)

// SetupRoutes konfiguriert alle API-Routen fuer den Authentifizierungs-Service.
// Routen sind in oeffentliche (ohne Auth) und geschuetzte (mit JWT) Gruppen unterteilt.
func SetupRoutes(router *gin.Engine, handler *AuthHandler, tokenGenerator ports.TokenGenerator) {
	// Globale Middleware anwenden
	router.Use(CORSMiddleware())

	// Health-Check Endpunkt (oeffentlich)
	router.GET("/health", handler.HealthCheck)

	// Authentifizierungs-Routen Gruppe
	auth := router.Group("/auth")
	{
		// Oeffentliche Routen (keine Authentifizierung erforderlich)
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)

		// Geschuetzte Routen (JWT-Authentifizierung erforderlich)
		authMiddleware := NewAuthMiddleware(tokenGenerator)
		protected := auth.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/me", handler.GetCurrentUser)
		}
	}

	// Legacy-Routen fuer Rueckwaertskompatibilitaet
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
}
