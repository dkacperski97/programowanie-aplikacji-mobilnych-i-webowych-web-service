package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
)

func index(w http.ResponseWriter, req *http.Request) {
	tmp := template.Must(template.ParseFiles("templates/index.html"))
	err := tmp.Execute(w, nil)
 
    if err != nil {
        panic(err)
    }
}

func signUpSender(w http.ResponseWriter, req *http.Request) {
	tmp := template.Must(template.ParseFiles("templates/signUpSender.html"))
	err := tmp.Execute(w, nil)
 
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
		Addr: ":8080",
		Handler: nil,
	}

	log.Println("Listening on :8080...")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}