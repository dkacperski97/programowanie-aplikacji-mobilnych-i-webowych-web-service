package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/project/handlers"
	"example.com/project/helpers"
	models "github.com/dkacperski97/programowanie-aplikacji-mobilnych-i-webowych-models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pmoule/go2hal/hal"
)

var client *redis.Client

func getLabels(w http.ResponseWriter, req *http.Request) {
	claims, _ := handlers.GetClaims(req.Context())

	var labels []models.Label
	var err error

	if claims.Role == "sender" {
		labels, err = helpers.GetLabelsBySender(client, claims.User)
	} else {
		labels, err = helpers.GetLabelsBySender(client, claims.User)
	}

	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	root := hal.NewResourceObject()

	self, _ := hal.NewLinkRelation("self")
	self.SetLink(&hal.LinkObject{Href: "/labels"})
	root.AddLink(self)

	var embeddedLabels []hal.Resource

	for _, label := range labels {
		href := fmt.Sprintf("/labels/%s", label.ID)
		self, _ := hal.NewLinkRelation("self")
		self.SetLink(&hal.LinkObject{Href: href})

		embeddedLabel := hal.NewResourceObject()
		embeddedLabel.AddLink(self)
		embeddedLabel.AddData(label)
		embeddedLabels = append(embeddedLabels, embeddedLabel)
	}

	labelsRelation, _ := hal.NewResourceRelation("labels")
	labelsRelation.SetResources(embeddedLabels)

	root.AddResource(labelsRelation)

	bytes, err := hal.NewEncoder().ToJSON(root)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/hal+json")
	w.Write(bytes)
}

func createLabel(w http.ResponseWriter, req *http.Request) {
	claims, _ := handlers.GetClaims(req.Context())

	if claims.Role != "sender" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	headerContentType := req.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		http.Error(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	var label *models.Label

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(label)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if label.Sender != claims.User {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = helpers.SaveLabel(client, label)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", "/labels/"+string(label.ID))
}

func removeLabel(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	labelID := vars["labelId"]

	claims, _ := handlers.GetClaims(req.Context())

	if claims.Role != "sender" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := helpers.RemoveLabel(
		client,
		claims.User,
		labelID,
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func index(w http.ResponseWriter, req *http.Request) {
	_, ok := handlers.GetClaims(req.Context())

	root := hal.NewResourceObject()

	self, _ := hal.NewLinkRelation("self")
	self.SetLink(&hal.LinkObject{Href: "/"})
	root.AddLink(self)

	if ok {
		labels, _ := hal.NewLinkRelation("labels")
		labels.SetLink(&hal.LinkObject{Href: "/labels"})
		root.AddLink(labels)
	}

	bytes, err := hal.NewEncoder().ToJSON(root)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/hal+json")
	w.Write(bytes)
}

func getRedisClient() *redis.Client {
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPass := os.Getenv("REDIS_PASS")
	redisDbString := os.Getenv("REDIS_DB")
	redisDb, err := strconv.Atoi(redisDbString)
	if err != nil {
		redisDb = 0
	}
	return redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPass,
		DB:       redisDb,
	})
}

func main() {
	err := godotenv.Load()
	if err == nil {
		log.Print(".env file loaded")
	}

	client = getRedisClient()
	defer client.Close()

	gob.Register(&time.Time{})

	mainRouter := mux.NewRouter()
	mainRouter.Handle("/", handlers.JwtHandler([]byte(os.Getenv("JWT_SECRET")), http.HandlerFunc(index), false)).Methods(http.MethodGet, http.MethodOptions)
	r := mainRouter.PathPrefix("/").Subrouter()
	r.Handle("/labels", http.HandlerFunc(getLabels)).Methods(http.MethodGet)
	r.Handle("/labels", http.HandlerFunc(createLabel)).Methods(http.MethodPost)
	r.Handle("/labels/{labelId}", http.HandlerFunc(removeLabel)).Methods(http.MethodDelete)
	// mainRouter.PathPrefix("/").Handler(handlers.JwtHandler([]byte(os.Getenv("JWT_SECRET")), r, true)).Methods(http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions)

	mainRouter.Use(mux.CORSMethodMiddleware(mainRouter))
	mainRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})
	http.Handle("/", mainRouter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	s := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	log.Println("Listening on :" + port + " ...")
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
