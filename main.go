package main

import (
	"os"
	"log"
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
)

func index(w http.ResponseWriter, req *http.Request) {
	tmp := template.New("_func").Funcs(template.FuncMap{
        "getDate": func() time.Time {
            return time.Now()
		},
	})
	tmp = template.Must(tmp.ParseGlob("templates/*.html"))
	err := tmp.ExecuteTemplate(w, "index.html", nil)
 
    if err != nil {
        panic(err)
    }
}

func signUpSender(w http.ResponseWriter, req *http.Request) {
	tmp := template.New("_func").Funcs(template.FuncMap{
        "getDate": func() time.Time {
            return time.Now()
		},
	})
	tmp = template.Must(tmp.ParseGlob("templates/*.html"))
	err := tmp.ExecuteTemplate(w, "signUpSender.html", nil)
 
    if err != nil {
        panic(err)
    }
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/sender/sign-up", signUpSender)
    http.Handle("/", r)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	s := &http.Server{
		Addr: ":" + port,
		Handler: nil,
	}

	log.Println("Listening on :" + port + "...")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}