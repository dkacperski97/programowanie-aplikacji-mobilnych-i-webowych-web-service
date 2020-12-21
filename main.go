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
	"example.com/project/models"
	sharedModels "github.com/dkacperski97/programowanie-aplikacji-mobilnych-i-webowych-models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pmoule/go2hal/hal"
)

var client *redis.Client

func getLabels(w http.ResponseWriter, req *http.Request) {
	claims, _ := handlers.GetClaims(req.Context())

	var labels []sharedModels.Label
	var err error

	if claims.Role == "sender" {
		labels, err = helpers.GetLabelsBySender(client, claims.User)
	} else {
		labels, err = helpers.GetLabels(client)
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
		embeddedLabel := hal.NewResourceObject()

		if label.AssignedParcel == "" {
			if claims.Role == "sender" {
				href := fmt.Sprintf("/labels/%s", label.ID)
				self, _ := hal.NewLinkRelation("self")
				self.SetLink(&hal.LinkObject{Href: href})
				embeddedLabel.AddLink(self)
			} else {
				assign, _ := hal.NewLinkRelation("assign")
				assign.SetLink(&hal.LinkObject{Href: "/parcels"})
				embeddedLabel.AddLink(assign)
			}
		}

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

	var label sharedModels.Label

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&label)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if label.Sender != claims.User {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = helpers.SaveLabel(client, &label)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func index(w http.ResponseWriter, req *http.Request) {
	claims, ok := handlers.GetClaims(req.Context())

	root := hal.NewResourceObject()

	self, _ := hal.NewLinkRelation("self")
	self.SetLink(&hal.LinkObject{Href: "/"})
	root.AddLink(self)

	if ok {
		labels, _ := hal.NewLinkRelation("labels")
		labels.SetLink(&hal.LinkObject{Href: "/labels"})
		root.AddLink(labels)

		if claims.Role != "sender" {
			parcels, _ := hal.NewLinkRelation("parcels")
			parcels.SetLink(&hal.LinkObject{Href: "/parcels"})
			root.AddLink(parcels)
		}
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

func getParcels(w http.ResponseWriter, req *http.Request) {
	claims, _ := handlers.GetClaims(req.Context())

	var parcels []models.Parcel
	var err error

	if claims.Role == "sender" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	parcels, err = models.GetParcels(client)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	root := hal.NewResourceObject()

	self, _ := hal.NewLinkRelation("self")
	self.SetLink(&hal.LinkObject{Href: "/parcels"})
	root.AddLink(self)

	var embeddedParcels []hal.Resource

	for _, parcel := range parcels {
		href := fmt.Sprintf("/parcels/%s", parcel.ID)
		self, _ := hal.NewLinkRelation("self")
		self.SetLink(&hal.LinkObject{Href: href})

		embeddedParcel := hal.NewResourceObject()
		embeddedParcel.AddLink(self)
		embeddedParcel.AddData(parcel)
		embeddedParcels = append(embeddedParcels, embeddedParcel)
	}

	parcelsRelation, _ := hal.NewResourceRelation("parcels")
	parcelsRelation.SetResources(embeddedParcels)

	root.AddResource(parcelsRelation)

	bytes, err := hal.NewEncoder().ToJSON(root)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/hal+json")
	w.Write(bytes)
}

func createParcel(w http.ResponseWriter, req *http.Request) {
	claims, _ := handlers.GetClaims(req.Context())

	if claims.Role == "sender" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	headerContentType := req.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		http.Error(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	type Label struct {
		ID sharedModels.LabelID `json:"labelId"`
	}
	var label Label

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&label)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	parcel, validationErr, err := models.CreateParcel(string(label.ID), models.ParcelStatusOnTheWay)
	if validationErr != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = parcel.Save(client)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", "/parcels/"+string(parcel.ID))
}

func changeParcelStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	parcelID := vars["parcelId"]

	claims, _ := handlers.GetClaims(req.Context())

	if claims.Role == "sender" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	headerContentType := req.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		http.Error(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	var parcel models.Parcel

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&parcel)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	parcel.ID = parcelID

	validationErr, err := models.IsParcelValid(parcel.ID, parcel.Status)
	if validationErr != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = parcel.UpdateStatus(client)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	r2 := mainRouter.PathPrefix("/").Subrouter()
	r2.HandleFunc("/labels", http.HandlerFunc(getLabels)).Methods(http.MethodGet, http.MethodOptions)
	r2.HandleFunc("/labels", http.HandlerFunc(createLabel)).Methods(http.MethodPost, http.MethodOptions)
	r2.HandleFunc("/labels/{labelId}", http.HandlerFunc(removeLabel)).Methods(http.MethodDelete, http.MethodOptions)
	r2.HandleFunc("/parcels", http.HandlerFunc(getParcels)).Methods(http.MethodGet, http.MethodOptions)
	r2.HandleFunc("/parcels", http.HandlerFunc(createParcel)).Methods(http.MethodPost, http.MethodOptions)
	r2.HandleFunc("/parcels/{parcelId}", http.HandlerFunc(changeParcelStatus)).Methods(http.MethodPut, http.MethodOptions)
	r2.Use(mux.CORSMethodMiddleware(r2))
	r2.Use(handlers.JwtHandler([]byte(os.Getenv("JWT_SECRET")), true))
	r2.Use(handlers.HeadersHandler())

	r := mainRouter.PathPrefix("/").Subrouter()
	r.HandleFunc("/", http.HandlerFunc(index)).Methods(http.MethodGet, http.MethodOptions)
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(handlers.JwtHandler([]byte(os.Getenv("JWT_SECRET")), false))
	r.Use(handlers.HeadersHandler())

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
