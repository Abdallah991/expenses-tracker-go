package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in our JWT tokens
type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

var (
	jwtSecret     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
)

// InitJWT initializes JWT configuration from environment variables
func InitJWT() error {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}

	// Parse access token expiry (default: 15 minutes)
	accessExpiryStr := os.Getenv("JWT_ACCESS_EXPIRY")
	if accessExpiryStr == "" {
		accessExpiry = 15 * time.Minute
	} else {
		var err error
		accessExpiry, err = time.ParseDuration(accessExpiryStr)
		if err != nil {
			return fmt.Errorf("invalid JWT_ACCESS_EXPIRY format: %w", err)
		}
	}

	// Parse refresh token expiry (default: 7 days)
	refreshExpiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshExpiryStr == "" {
		refreshExpiry = 7 * 24 * time.Hour
	} else {
		var err error
		refreshExpiry, err = time.ParseDuration(refreshExpiryStr)
		if err != nil {
			return fmt.Errorf("invalid JWT_REFRESH_EXPIRY format: %w", err)
		}
	}

	return nil
}

// GenerateAccessToken generates a new access token for a user
func GenerateAccessToken(userID int, email string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "expenses-tracker-go",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	// signing method is HMAC-SHA265
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// GenerateRefreshToken generates a new refresh token (just a random string, not JWT)
func GenerateRefreshToken() (string, error) {
	// Generate a cryptographically secure random token
	// This will be stored in the database and used to generate new access tokens
	return GenerateSecureToken(32)
}

// ValidateAccessToken validates and parses an access token
func ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}
	// return JWT claims if signing method correct & compliant with cliams
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractTokenFromHeader extracts the Bearer token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check if it starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token cannot be empty")
	}

	return token, nil
}

// GetTokenExpiry returns the expiry time for access tokens
func GetTokenExpiry() time.Duration {
	return accessExpiry
}

// GetRefreshTokenExpiry returns the expiry time for refresh tokens
func GetRefreshTokenExpiry() time.Duration {
	return refreshExpiry
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
