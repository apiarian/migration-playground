package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var address string

func init() {
	flag.StringVar(&address, "address", "127.0.0.1:3000", "address and port on which to listen for HTTP requests")
}

func main() {
	flag.Parse()

	ts := NewMemoryThings()

	r := mux.NewRouter()

	t := r.PathPrefix("/things").Subrouter()
	t.HandleFunc("/", MakeListThingsHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/", MakeCreateThingHandler(ts)).Methods(http.MethodPost)
	t.HandleFunc("/{id}", MakeGetThingHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/{id}", MakeUpdateThingHandlerFunc(ts)).Methods(http.MethodPost)

	http.Handle("/", r)

	log.Print("listening on ", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
