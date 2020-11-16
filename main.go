package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/rbcervilla/redisstore/v8"
	"golang.org/x/crypto/bcrypt"
)

var (
	client      *redis.Client
	store       *redisstore.RedisStore
	sessionName string
)

type User struct {
	Login        string
	PasswordHash []byte
	Email        string
	Firstname    string
	Lastname     string
	Address      string
}

func getUser(login, password, email, firstname, lastname, address string) *User {
	u := new(User)
	u.Login = login
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	u.PasswordHash = passwordHash
	u.Email = email
	u.Firstname = firstname
	u.Lastname = lastname
	u.Address = address
	return u
}

func getTemplates() *template.Template {
	tmp := template.New("_func").Funcs(template.FuncMap{
		"getDate": time.Now,
	})
	tmp = template.Must(tmp.ParseGlob("templates/*.html"))
	return tmp
}

func index(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates()
	err := tmp.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		panic(err)
	}
}

func getRegisterSender(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates()
	err := tmp.ExecuteTemplate(w, "signUpSender.html", nil)
	if err != nil {
		panic(err)
	}
}

func postRegisterSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	user := getUser(
		req.Form.Get("login"),
		req.Form.Get("password"),
		req.Form.Get("email"),
		req.Form.Get("firstname"),
		req.Form.Get("lastname"),
		req.Form.Get("address"),
	)
	saveUser(user)

	tmp := getTemplates()
	err = tmp.ExecuteTemplate(w, "signUpSender.html", nil)
	if err != nil {
		panic(err)
	}
}

func getLoginSender(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates()
	err := tmp.ExecuteTemplate(w, "loginSender.html", nil)
	if err != nil {
		panic(err)
	}
}

func postLoginSender(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	isValid := verifyUser(req.Form.Get("login"), req.Form.Get("password"))

	if !isValid {
		tmp := getTemplates()
		err = tmp.ExecuteTemplate(w, "loginSender.html", nil)
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
	session, err := store.Get(req, sessionName)
	if err != nil {
		log.Fatal("Failed getting session: ", err)
	}

	if session.IsNew {
		// w.WriteHeader(http.StatusForbidden)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	tmp := getTemplates()
	err = tmp.ExecuteTemplate(w, "dashboard.html", nil)
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

func verifyUser(login, password string) bool {
	hash, err := client.HGet(context.Background(), "user:"+login, "passwordHash").Bytes()
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}

func saveUser(user *User) {
	client.HSet(context.Background(), "user:"+user.Login, map[string]interface{}{
		"passwordHash": user.PasswordHash,
		"email":        user.Email,
		"firstname":    user.Firstname,
		"lastname":     user.Lastname,
		"address":      user.Address,
	})
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
	r.HandleFunc("/sender/dashboard", showDashboard)
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
