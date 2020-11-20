package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/project/handlers"
	"example.com/project/models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/rbcervilla/redisstore/v8"
)

var (
	client      *redis.Client
	store       *redisstore.RedisStore
	sessionName string
)

func getTemplates(req *http.Request) *template.Template {
	session, err := store.Get(req, sessionName)
	tmp := template.New("_func").Funcs(template.FuncMap{
		"getDate": time.Now,
		"getSession": func() *sessions.Session {
			if err != nil || session.IsNew {
				return nil
			}
			return session
		},
	})
	tmp = template.Must(tmp.ParseGlob("templates/*.html"))
	return tmp
}

func index(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
	}
}

type registerSenderPageData struct {
	Error error
}

func getRegisterSender(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "signUpSender.html", &registerSenderPageData{
		Error: nil,
	})
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
	}
}

func postRegisterSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	user, validationErr, err := models.CreateUser(
		req.Form.Get("login"),
		req.Form.Get("password"),
		req.Form.Get("email"),
		req.Form.Get("firstname"),
		req.Form.Get("lastname"),
		req.Form.Get("address"),
	)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	if validationErr != nil {
		tmp := getTemplates(req)
		err = tmp.ExecuteTemplate(w, "signUpSender.html", &registerSenderPageData{
			Error: validationErr,
		})
		if err != nil {
			handleError(w, req, http.StatusInternalServerError)
		}
		return
	}
	user.Save(client)
	http.Redirect(w, req, "/sender/login", http.StatusSeeOther)
}

type loginSenderPageData struct {
	Error error
}

func getLoginSender(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "loginSender.html", &registerSenderPageData{
		Error: nil,
	})
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
	}
}

func postLoginSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	isValid, err := models.Verify(client, req.Form.Get("login"), req.Form.Get("password"))
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	if !isValid {
		tmp := getTemplates(req)
		err = tmp.ExecuteTemplate(w, "loginSender.html", &registerSenderPageData{
			Error: errors.New("Niepoprawne dane logowania"),
		})
		if err != nil {
			handleError(w, req, http.StatusInternalServerError)
		}
		return
	}

	session, err := store.Get(req, sessionName)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	session.Values["user"] = req.Form.Get("login")
	session.Values["loginTime"] = time.Now()

	if err = sessions.Save(req, w); err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func logoutSender(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1

	if err = sessions.Save(req, w); err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

type showDashboardPageData struct {
	Labels []models.Label
}

func showDashboard(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	sender, exists := session.Values["user"]
	if exists == false {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	labels, err := models.GetLabelsBySender(client, sender.(string))
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	tmp := getTemplates(req)
	err = tmp.ExecuteTemplate(w, "dashboard.html", &showDashboardPageData{
		Labels: labels,
	})
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
	}
}

type createLabelPageData struct {
	Error error
}

func getCreateLabel(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "createLabel.html", &createLabelPageData{
		Error: nil,
	})
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
}

func postCreateLabel(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	session, err := store.Get(req, sessionName)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	sender, exists := session.Values["user"]
	if exists == false {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	label, validationErr, err := models.CreateLabel(
		sender.(string),
		req.Form.Get("recipient"),
		req.Form.Get("locker"),
		req.Form.Get("size"),
	)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	if validationErr != nil {
		tmp := getTemplates(req)
		err = tmp.ExecuteTemplate(w, "createLabel.html", &createLabelPageData{
			Error: validationErr,
		})
		if err != nil {
			handleError(w, req, http.StatusInternalServerError)
		}
		return
	}
	label.Save(client)
	http.Redirect(w, req, "/sender/dashboard", http.StatusSeeOther)
}

func removeLabel(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	labelID := vars["labelId"]

	session, err := store.Get(req, sessionName)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	sender, exists := session.Values["user"]
	if exists == false {
		handleError(w, req, http.StatusInternalServerError)
		return
	}

	err = models.RemoveLabel(
		client,
		sender.(string),
		labelID,
	)
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, "/sender/dashboard", http.StatusSeeOther)
}

func checkAvailability(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	login := vars["login"]

	numberOfKeys, err := client.Exists(context.Background(), "user:"+login).Uint64()
	if err != nil {
		handleError(w, req, http.StatusInternalServerError)
		return
	}
	loginAvailability := "available"
	if numberOfKeys != 0 {
		loginAvailability = "taken"
	}
	data := map[string]interface{}{
		login: loginAvailability,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

type handleErrorPageData struct {
	StatusCode int
	StatusText string
}

func handleError(w http.ResponseWriter, req *http.Request, code int) {
	w.WriteHeader(code)
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "error.html", &handleErrorPageData{
		StatusCode: code,
		StatusText: http.StatusText(code),
	})
	if err != nil {
		fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
		return
	}
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
	return redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       0,
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}

	client = getRedisClient()
	defer client.Close()

	store, err = redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		log.Print("Failed to create redis store: ", err)
	}
	var secureCookie bool
	if os.Getenv("SECURE_COOKIE") == "" || os.Getenv("SECURE_COOKIE") == "FALSE" {
		secureCookie = false
	} else {
		secureCookie = true
	}
	store.Options(sessions.Options{
		Secure:   secureCookie,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   60 * 60 * 24,
	})
	sessionName = os.Getenv("SESSION_NAME")
	if sessionName == "" {
		sessionName = "session"
	}
	gob.Register(&time.Time{})

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.Handle("/sender/register", handlers.WithoutSessionHandler(store, sessionName, http.HandlerFunc(getRegisterSender), handleError)).Methods("GET")
	r.Handle("/sender/register", handlers.WithoutSessionHandler(store, sessionName, http.HandlerFunc(postRegisterSender), handleError)).Methods("POST")
	r.Handle("/sender/login", handlers.WithoutSessionHandler(store, sessionName, http.HandlerFunc(getLoginSender), handleError)).Methods("GET")
	r.Handle("/sender/login", handlers.WithoutSessionHandler(store, sessionName, http.HandlerFunc(postLoginSender), handleError)).Methods("POST")
	r.Handle("/sender/logout", handlers.SessionHandler(store, sessionName, http.HandlerFunc(logoutSender), handleError))
	r.Handle("/sender/dashboard", handlers.SessionHandler(store, sessionName, http.HandlerFunc(showDashboard), handleError))
	r.Handle("/sender/labels/create", handlers.SessionHandler(store, sessionName, http.HandlerFunc(getCreateLabel), handleError)).Methods("GET")
	r.Handle("/sender/labels/create", handlers.SessionHandler(store, sessionName, http.HandlerFunc(postCreateLabel), handleError)).Methods("POST")
	r.Handle("/sender/labels/{labelId}/remove", handlers.SessionHandler(store, sessionName, http.HandlerFunc(removeLabel), handleError)).Methods("POST")
	r.HandleFunc("/check/{login}", checkAvailability)
	http.Handle("/", r)

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
