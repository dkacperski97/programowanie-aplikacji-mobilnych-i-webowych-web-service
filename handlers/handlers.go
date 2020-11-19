package handlers

import (
	"log"
	"net/http"

	"github.com/rbcervilla/redisstore/v8"
)

type sessionHandler struct {
	store       *redisstore.RedisStore
	sessionName string
	handler     http.Handler
}

func (h sessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	session, err := h.store.Get(req, h.sessionName)
	if err != nil {
		log.Fatal("Failed getting session: ", err)
	}

	if session.IsNew {
		// w.WriteHeader(http.StatusForbidden)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	h.handler.ServeHTTP(w, req)
}

// SessionHandler return a http.Handler that wraps h and checks if the session is available
func SessionHandler(store *redisstore.RedisStore, sessionName string, h http.Handler) http.Handler {
	return sessionHandler{store, sessionName, h}
}
