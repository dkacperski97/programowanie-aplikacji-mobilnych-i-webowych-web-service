package main

import (
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

	s := &http.Server{
		Addr: ":5000",
		Handler: nil,
	}

	log.Println("Listening on :5000...")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}