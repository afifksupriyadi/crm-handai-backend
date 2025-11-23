package security

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// supportedJWTSignAlgorithm defines the allowed JWT signing algorithms.
// This map is used to validate the signing algorithm during token generation.
var supportedJWTSignAlgorithm = map[string]bool{
	"HS256": true,
	"RS256": true,
}

// Sentinel errors for token validation.
var (
	ErrInvalidInput  = errors.New("invalid input parameters")
	ErrTokenExpired  = errors.New("token has expired")
	ErrInvalidIssuer = errors.New("invalid token issuer")
)

// Claims represents the custom claims structure for JWT tokens.
type Claims struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Exp     time.Time
	Iat     time.Time
	Iss     string
}

// GenerateTokenParam holds the parameters required to generate a JWT token.
type GenerateTokenParam struct {
	// From config
	Expiration time.Duration
	Version    int
	Issuer     string
	Algorithm  string
	SecretKey  string
	// User data
	UserID int
	Name   string
}

func (p *GenerateTokenParam) validate() error {
	if p.UserID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", ErrInvalidInput)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}
	if !supportedJWTSignAlgorithm[p.Algorithm] {
		return fmt.Errorf("%w: unsupported signing algorithm: %s", ErrInvalidInput, p.Algorithm)
	}
	return nil
}

// GenerateUserToken creates a signed JWT for the given user.
func GenerateUserToken(p GenerateTokenParam) (string, error) {
	if err := p.validate(); err != nil {
		return "", err
	}

	now := time.Now()
	exp := now.Add(p.Expiration)

	token, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).
		Issuer(p.Issuer).
		IssuedAt(now).
		NotBefore(now).
		Expiration(exp).
		Subject(strconv.Itoa(p.UserID)).
		Claim("name", p.Name).
		Claim("version", p.Version).
		Build()
	if err != nil {
		return "", fmt.Errorf("failed to build token: %w", err)
	}

	signAlg := jwa.SignatureAlgorithm(p.Algorithm)

	signed, err := jwt.Sign(token, jwt.WithKey(signAlg, []byte(p.SecretKey)))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return string(signed), nil
}

// ValidateToken verifies the token using secret, algorithm, and issuer.
// Returns parsed token or detailed error.
func ValidateToken(tokenString, secretKey, expectedIssuer, algorithm string) (jwt.Token, error) {
	signAlg := jwa.SignatureAlgorithm(algorithm)

	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKey(signAlg, []byte(secretKey)),
		jwt.WithValidate(true),
		jwt.WithIssuer(expectedIssuer),
	)

	if err != nil {
		if strings.Contains(err.Error(), "\"exp\" not satisfied") {
			return nil, ErrTokenExpired
		}
		if strings.Contains(err.Error(), "\"iss\" not satisfied") {
			return nil, ErrInvalidIssuer
		}
		return nil, fmt.Errorf("failed to parse and validate token: %w", err)
	}

	if token == nil {
		return nil, errors.New("token is nil after parsing")
	}

	return token, nil
}

// ExtractClaims extracts custom claims from a JWT token.
func ExtractClaims(token jwt.Token) *Claims {
	claims := &Claims{
		Exp: token.Expiration(),
		Iat: token.IssuedAt(),
		Iss: token.Issuer(),
	}

	claims.UserID, _ = strconv.Atoi(token.Subject())

	if v, ok := token.PrivateClaims()["name"].(string); ok {
		claims.Name = v
	}
	if v, ok := token.PrivateClaims()["version"].(float64); ok {
		claims.Version = int(v)
	}

	return claims
}
