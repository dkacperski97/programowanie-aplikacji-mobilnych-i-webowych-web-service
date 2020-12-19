package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	models "github.com/dkacperski97/programowanie-aplikacji-mobilnych-i-webowych-models"
)

type (
	tokenKey   string
	jwtHandler struct {
		secret          []byte
		handler         http.Handler
		isTokenRequired bool
	}
)

const key tokenKey = "ParsedToken"

func (h jwtHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	tokenString := strings.Replace(req.Header.Get("Authorization"), "Bearer ", "", 1)

	token, err := jwt.ParseWithClaims(tokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return h.secret, nil
	})

	if err == nil {
		if claims, ok := token.Claims.(*models.UserClaims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), key, claims)
			h.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}

	if !h.isTokenRequired {
		h.handler.ServeHTTP(w, req)
		return
	}

	http.Error(w, "Incorrect authorization header", http.StatusBadRequest)
}

func GetClaims(ctx context.Context) (*models.UserClaims, bool) {
	claims, ok := ctx.Value(key).(*models.UserClaims)
	return claims, ok
}

// JwtHandler return a http.Handler that wraps h and checks if the jwt token is correct
func JwtHandler(secret []byte, h http.Handler, isTokenRequired bool) http.Handler {
	return jwtHandler{secret, h, isTokenRequired}
}
