package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// TokenLifetime is how long a login token stays valid (§1.2).
const TokenLifetime = 8 * time.Hour

// AuthService issues and verifies login tokens.
type AuthService struct {
	db        *gorm.DB
	secret    []byte
	hashCost  int
	now       func() time.Time
	tokenLife time.Duration
}

// Claims is the JWT payload. Subject carries the user id.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthService builds an AuthService signing with secret.
func NewAuthService(db *gorm.DB, secret string) *AuthService {
	return &AuthService{
		db:        db,
		secret:    []byte(secret),
		hashCost:  bcrypt.DefaultCost,
		now:       func() time.Time { return time.Now().UTC() },
		tokenLife: TokenLifetime,
	}
}

// SetHashCost lowers the bcrypt cost; tests use it to stay fast.
func (s *AuthService) SetHashCost(cost int) {
	s.hashCost = cost
}

// HashPassword returns the bcrypt hash to store for a password.
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.hashCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Login checks credentials and returns a signed token with the user (§1.2).
// A soft-deleted user is invisible to the default scope, so they fail the
// lookup; a deactivated one is refused explicitly.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, unauthorized("invalid email or password")
	}
	if err != nil {
		return "", nil, fmt.Errorf("look up user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", nil, unauthorized("invalid email or password")
	}
	if !user.IsActive {
		return "", nil, unauthorized("account is deactivated")
	}
	token, err := s.issue(user)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func (s *AuthService) issue(user models.User) (string, error) {
	now := s.now()
	claims := Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenLife)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Authenticate verifies a bearer token and returns the live user it names.
// The user is re-read on every request so deactivation and deletion take
// effect immediately rather than when the token expires.
func (s *AuthService) Authenticate(ctx context.Context, token string) (*models.User, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		return nil, unauthorized("invalid or expired token")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return nil, unauthorized("invalid token subject")
	}
	var user models.User
	err = s.db.WithContext(ctx).First(&user, uint(id)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, unauthorized("user no longer exists")
	}
	if err != nil {
		return nil, fmt.Errorf("look up user: %w", err)
	}
	if !user.IsActive {
		return nil, unauthorized("account is deactivated")
	}
	return &user, nil
}
