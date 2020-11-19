package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/project/auth"
	"example.com/project/handlers"
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
	if err != nil {
		log.Fatal("Failed getting session: ", err)
	}

	tmp := template.New("_func").Funcs(template.FuncMap{
		"getDate": time.Now,
		"getSession": func() *sessions.Session {
			if session.IsNew {
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
		panic(err)
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
		panic(err)
	}
}

func postRegisterSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	user, validationErr, err := auth.CreateUser(
		req.Form.Get("login"),
		req.Form.Get("password"),
		req.Form.Get("email"),
		req.Form.Get("firstname"),
		req.Form.Get("lastname"),
		req.Form.Get("address"),
	)
	if err != nil {
		panic(err)
	}
	if validationErr != nil {
		tmp := getTemplates(req)
		err = tmp.ExecuteTemplate(w, "signUpSender.html", &registerSenderPageData{
			Error: validationErr,
		})
		if err != nil {
			panic(err)
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
		panic(err)
	}
}

func postLoginSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	isValid := auth.Verify(client, req.Form.Get("login"), req.Form.Get("password"))

	if !isValid {
		tmp := getTemplates(req)
		err = tmp.ExecuteTemplate(w, "loginSender.html", &registerSenderPageData{
			Error: errors.New("Niepoprawne dane logowania"),
		})
		if err != nil {
			panic(err)
		}
		return
	}

	session, err := store.Get(req, sessionName)
	if err != nil {
		log.Fatal("Failed getting session: ", err)
	}

	session.Values["loginTime"] = time.Now()

	if err = sessions.Save(req, w); err != nil {
		log.Fatal("Failed saving session: ", err)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func logoutSender(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		log.Fatal("Failed getting session: ", err)
	}

	session.Options.MaxAge = -1

	if err = sessions.Save(req, w); err != nil {
		log.Fatal("Failed deleting session: ", err)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func showDashboard(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "dashboard.html", nil)
	if err != nil {
		panic(err)
	}
}

func getCreateLabel(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "createLabel.html", nil)
	if err != nil {
		panic(err)
	}
}

func postCreateLabel(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates(req)
	err := tmp.ExecuteTemplate(w, "createLabel.html", nil)
	if err != nil {
		panic(err)
	}
}

func checkAvailability(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	login := vars["login"]

	numberOfKeys, err := client.Exists(context.Background(), "user:"+login).Uint64()
	if err != nil {
		panic(err)
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
		log.Fatal("Error loading .env file")
	}

	client = getRedisClient()
	defer client.Close()

	store, err = redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		log.Fatal("Failed to create redis store: ", err)
	}
	sessionName = os.Getenv("SESSION_NAME")
	if sessionName == "" {
		sessionName = "session"
	}
	gob.Register(&time.Time{})

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/sender/register", getRegisterSender).Methods("GET")
	r.HandleFunc("/sender/register", postRegisterSender).Methods("POST")
	r.HandleFunc("/sender/login", getLoginSender).Methods("GET")
	r.HandleFunc("/sender/login", postLoginSender).Methods("POST")
	r.HandleFunc("/sender/logout", logoutSender)
	r.Handle("/sender/dashboard", handlers.SessionHandler(store, sessionName, http.HandlerFunc(showDashboard)))
	r.Handle("/sender/labels/create", handlers.SessionHandler(store, sessionName, http.HandlerFunc(getCreateLabel))).Methods("GET")
	r.Handle("/sender/labels/create", handlers.SessionHandler(store, sessionName, http.HandlerFunc(postCreateLabel))).Methods("POST")
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
