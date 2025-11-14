package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtSecret     []byte
	jwtExpiration time.Duration
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(jwtSecret string, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:     []byte(jwtSecret),
		jwtExpiration: jwtExpiration,
	}
}

// HashPassword hashes a password using bcrypt
func (a *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func (a *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateAccessToken generates a JWT access token
func (a *AuthService) GenerateAccessToken(userID int64, email string) (string, error) {
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "yourdht",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// GenerateRefreshToken generates a refresh token
func (a *AuthService) GenerateRefreshToken(userID int64) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "yourdht",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateAccessToken validates a JWT access token
func (a *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateAPIKey generates a random API key
func (a *AuthService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashAPIKey hashes an API key using bcrypt
func (a *AuthService) HashAPIKey(apiKey string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyAPIKey verifies an API key against a hash
func (a *AuthService) VerifyAPIKey(hashedKey, apiKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(apiKey))
}
