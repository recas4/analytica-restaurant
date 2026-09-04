package rest

import (
	"auth-service/internal/application/ports"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware stellt JWT-Authentifizierung fuer geschuetzte Routen bereit.
type AuthMiddleware struct {
	tokenGenerator ports.TokenGenerator
}

// NewAuthMiddleware erstellt eine neue Authentifizierungs-Middleware Instanz.
func NewAuthMiddleware(tokenGenerator ports.TokenGenerator) *AuthMiddleware {
	return &AuthMiddleware{
		tokenGenerator: tokenGenerator,
	}
}

// RequireAuth validiert das JWT-Token und extrahiert Benutzerinformationen.
// Geschuetzte Routen sollten diese Middleware verwenden um Authentifizierung sicherzustellen.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Authorization header is required",
			})
			return
		}

		// Token aus dem "Bearer <token>" Format extrahieren
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_format",
				"message": "Authorization header must be in 'Bearer <token>' format",
			})
			return
		}

		tokenString := parts[1]
		userID, err := m.tokenGenerator.Validate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token is invalid or expired",
			})
			return
		}

		// User-ID im Context speichern fuer nachfolgende Handler
		c.Set("user_id", userID)
		c.Next()
	}
}

// CORSMiddleware behandelt Cross-Origin Resource Sharing Konfiguration.
// Ermoeglicht API-Zugriff von verschiedenen Domains.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
