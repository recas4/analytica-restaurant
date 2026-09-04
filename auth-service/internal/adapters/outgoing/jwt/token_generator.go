package jwt

import (
	"auth-service/internal/application/ports"
	"auth-service/internal/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Fehlerdefinitionen fuer Token-Operationen.
var (
	ErrInvalidToken = errors.New("Ungueltiges oder abgelaufenes Token")
	ErrTokenParsing = errors.New("Token konnte nicht geparst werden")
)

// JWTTokenGenerator implementiert TokenGenerator mit JSON Web Tokens.
type JWTTokenGenerator struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTTokenGenerator erstellt einen neuen JWT-Generator mit der angegebenen Konfiguration.
func NewJWTTokenGenerator(secretKey string, expirySeconds int) ports.TokenGenerator {
	return &JWTTokenGenerator{
		secretKey:     []byte(secretKey),
		tokenDuration: time.Duration(expirySeconds) * time.Second,
	}
}

// Generate erstellt ein signiertes JWT-Token mit Benutzer-Claims.
func (g *JWTTokenGenerator) Generate(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(g.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secretKey)
}

// Validate parst und validiert das Token und gibt die User-ID zurueck wenn gueltig.
func (g *JWTTokenGenerator) Validate(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Signatur-Methode verifizieren
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return g.secretKey, nil
	})

	if err != nil {
		return "", ErrTokenParsing
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}
