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
	eh          func(w http.ResponseWriter, req *http.Request, code int)
}

func (h sessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	session, err := h.store.Get(req, h.sessionName)
	if err != nil {
		h.eh(w, req, http.StatusInternalServerError)
		log.Fatal("Failed getting session: ", err)
		return
	}

	if session.IsNew {
		http.Redirect(w, req, "/sender/login", http.StatusSeeOther)
		return
	}

	h.handler.ServeHTTP(w, req)
}

// SessionHandler return a http.Handler that wraps h and checks if the session is available
func SessionHandler(store *redisstore.RedisStore, sessionName string, h http.Handler, eh func(w http.ResponseWriter, req *http.Request, code int)) http.Handler {
	return sessionHandler{store, sessionName, h, eh}
}
