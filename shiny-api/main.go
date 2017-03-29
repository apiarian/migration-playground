package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var address string
var brokers string
var new_topic string

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
		&new_topic,
		"new-topic",
		fmt.Sprintf("thing-commands-%d", time.Now().Unix()),
		"the topic for tracking things for the shiny api",
	)
}

func main() {
	flag.Parse()

	kc, err := NewKafkaClient(strings.Split(brokers, ","), new_topic)
	if err != nil {
		log.Fatal("failed to create kafka client: ", err)
	}
	defer kc.Close()

	log.Print("commands are on ", new_topic)

	u := NewUpdater(kc, new_topic, true)
	errs := u.Start()

	ts := NewStreamThings(kc, u)

	r := mux.NewRouter()

	t := r.PathPrefix("/things").Subrouter()
	t.HandleFunc("/", MakeListThingsHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/", MakeCreateThingHandler(ts)).Methods(http.MethodPost)
	t.HandleFunc("/{id}", MakeGetThingHandlerFunc(ts)).Methods(http.MethodGet)
	t.HandleFunc("/{id}", MakeUpdateThingHandlerFunc(ts)).Methods(http.MethodPatch)

	r.HandleFunc("/commands/{id}", MakeCheckCommandHandler(ts)).Methods(http.MethodGet)

	http.Handle("/", r)

	log.Print("listening on ", address)
	s := &http.Server{
		Addr:    address,
		Handler: http.DefaultServeMux,
	}
	d := make(chan struct{})
	go func(s *http.Server, d chan<- struct{}) {
		log.Print("server l&s error: ", s.ListenAndServe())
		d <- struct{}{}
	}(s, d)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case <-signals:
		log.Print("got an interrupt")

	case err := <-errs:
		log.Print("updater start error: ", err)
	}

	err = s.Shutdown(context.Background())
	if err != nil {
		log.Print("server shutdown error: ", err)
	} else {
		log.Print("server shut down cleanly")
	}

	<-d
	log.Print("server goroutine finished")

	log.Print("really done now.")
}
