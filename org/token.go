package org

import (
	"errors"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/uptrace/bun-realworld-app/app"
)

func decodeUserToken(app *app.App, jwtToken string) (uint64, error) {
	if len(jwtToken) == 0 {
		return 0, errors.New("token is missing or empty")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.Config().SecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims := token.Claims.(*jwt.StandardClaims)

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func CreateUserToken(app *app.App, userID uint64, ttl time.Duration) (string, error) {
	claims := &jwt.StandardClaims{
		Subject:   strconv.FormatUint(userID, 10),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key := []byte(app.Config().SecretKey)
	return token.SignedString(key)
}
