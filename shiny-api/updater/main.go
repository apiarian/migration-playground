package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
)

var brokers string
var command_topic string

func init() {
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

	log.Print("working with commands on ", command_topic)
}
