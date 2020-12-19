package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	models "github.com/dkacperski97/programowanie-aplikacji-mobilnych-i-webowych-models"
	"github.com/gorilla/mux"
)

type (
	tokenKey string
)

const key tokenKey = "ParsedToken"

func GetClaims(ctx context.Context) (*models.UserClaims, bool) {
	claims, ok := ctx.Value(key).(*models.UserClaims)
	return claims, ok
}

func JwtHandler(secret []byte, isTokenRequired bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			tokenString := strings.Replace(req.Header.Get("Authorization"), "Bearer ", "", 1)

			token, err := jwt.ParseWithClaims(tokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return secret, nil
			})

			if err == nil {
				if claims, ok := token.Claims.(*models.UserClaims); ok && token.Valid {
					ctx := context.WithValue(req.Context(), key, claims)
					next.ServeHTTP(w, req.WithContext(ctx))
					return
				}
			}

			if !isTokenRequired {
				next.ServeHTTP(w, req)
				return
			}

			http.Error(w, "Incorrect authorization header", http.StatusBadRequest)
		})
	}
}
