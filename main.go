package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

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

func signUpSender(w http.ResponseWriter, req *http.Request) {
	tmp := getTemplates()
	err := tmp.ExecuteTemplate(w, "signUpSender.html", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/sender/register", signUpSender)
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
