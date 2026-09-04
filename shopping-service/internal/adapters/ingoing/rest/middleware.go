package rest

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Kontext-Schluessel fuer Benutzerinformationen
const (
	ContextUserIDKey   = "user_id"
	ContextUserRoleKey = "user_role"
)

// JWTMiddleware validiert JWT-Tokens und setzt Benutzerinformationen im Kontext.
func JWTMiddleware() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET nicht gesetzt")
	}

	return func(c *gin.Context) {
		tokenStr, ok := extractTokenFromHeader(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Fehlender oder ungueltiger Authorization-Header"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		}, jwt.WithLeeway(5*time.Second))

		if err != nil || !token.Valid {
			log.Printf("Token-Validierung fehlgeschlagen: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Ungueltiger Token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Ungueltige Token-Claims"})
			return
		}

		// Ablaufzeit pruefen
		if expVal, ok := claims["exp"]; ok {
			if v, ok := expVal.(float64); ok && int64(v) < time.Now().Unix() {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token abgelaufen"})
				return
			}
		}

		// Benutzer-ID extrahieren
		userID := extractUserID(claims)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_id Claim fehlt"})
			return
		}
		c.Set(ContextUserIDKey, userID)

		// Rolle extrahieren
		if role, ok := claims["role"].(string); ok && role != "" {
			c.Set(ContextUserRoleKey, role)
		}

		c.Next()
	}
}

// RequireRole prueft ob der Benutzer die erforderliche Rolle hat.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ContextUserRoleKey)
		if !ok || v != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unzureichende Berechtigung"})
			return
		}
		c.Next()
	}
}

// GetUserID holt die Benutzer-ID aus dem Kontext.
func GetUserID(c *gin.Context) (string, bool) {
	id, ok := c.Get(ContextUserIDKey)
	if !ok {
		return "", false
	}
	s, ok := id.(string)
	return s, ok
}

// extractTokenFromHeader extrahiert den Token aus dem Authorization-Header.
func extractTokenFromHeader(c *gin.Context) (string, bool) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", false
	}
	return parts[1], true
}

// extractUserID extrahiert die Benutzer-ID aus den JWT-Claims.
func extractUserID(claims jwt.MapClaims) string {
	if id, ok := claims["user_id"]; ok {
		return toString(id)
	}
	if sub, ok := claims["sub"]; ok {
		return toString(sub)
	}
	return ""
}

// toString konvertiert verschiedene Typen zu String.
func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}
