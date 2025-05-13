package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateRefreshToken() string {
	buf := make([]byte, 32)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

type JwtClaims struct {
	UserID uuid.UUID `json:"userID"`
	jwt.RegisteredClaims
}

func GenerateJWTAccessToken(claims JwtClaims) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(os.Getenv("SECRET")))
}

func ParseJWTTokenString(tokenString string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
