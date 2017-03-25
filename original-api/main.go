package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var address string
var brokers string
var original_topic string

func init() {
	flag.StringVar(
		&address,
		"address",
		"127.0.0.1:3000",
		"address and port on which to listen for HTTP requests",
	)
	flag.StringVar(
		&brokers,
		"brokers",
		"127.0.0.1:9092",
		"addresses of the kafka brokers to talk to",
	)
	flag.StringVar(
		&original_topic,
		"original-topic",
		fmt.Sprintf("things-%d", time.Now().Unix()),
		"the original topic on which things are published",
	)
}

func main() {
	flag.Parse()

	ts := NewMemoryThings()
	defer ts.Close()

	kc, err := NewKafkaClient(strings.Split(brokers, ","), original_topic)
	if err != nil {
		log.Fatal("failed to create kafka client: ", err)
	}
	defer kc.Close()

	go kc.PublishStream(ts.ThingStream())
	log.Print("publishing things to ", original_topic)

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
