package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
	"github.com/gorilla/mux"
)

var address string
var brokers string
var command_topic string

func init() {
	flag.StringVar(
		&address,
		"address",
		"127.0.0.1:9000",
		"address and port on which to listen for HTTP requests",
	)
	flag.StringVar(
		&brokers,
		"brokers",
		"127.0.0.1:9092",
		"addresses of the kafka brokers to talk to",
	)
	flag.StringVar(
		&command_topic,
		"command-topic",
		fmt.Sprintf("thing-commands-%d", time.Now().Unix()),
		"the command topic for coordinating thing creation and updates",
	)
}

func main() {
	flag.Parse()

	kc, err := common.NewKafkaClient(strings.Split(brokers, ","), command_topic)
	if err != nil {
		log.Fatal("failed to create kafka client: ", err)
	}
	defer kc.Close()

	log.Print("commands are on ", command_topic)

	ts := NewStreamThings(kc)

	r := mux.NewRouter()

	t := r.PathPrefix("/things").Subrouter()
	t.HandleFunc("/", MakeListThingsHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/", MakeCreateThingHandler(ts)).Methods(http.MethodPost)
	t.HandleFunc("/{id}", MakeGetThingHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/{id}", MakeUpdateThingHandlerFunc(ts)).Methods(http.MethodPatch)

	r.HandleFunc("/commands/{id}", MakeCheckCommandHandler(ts)).Methods(http.MethodGet)

	http.Handle("/", r)

	log.Print("listening on ", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
