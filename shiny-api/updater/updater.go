package main

import (
	"log"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
)

type Updater struct {
	kc            *common.KafkaClient
	command_topic string
	done          chan struct{}
}

func NewUpdater(kc *common.KafkaClient, command_topic string) *Updater {
	return &Updater{
		kc:            kc,
		command_topic: command_topic,
		done:          make(chan struct{}),
	}
}

func (u *Updater) Close() {
	close(u.done)
}

func (u *Updater) DoYourThing() error {
	return u.kc.RegisterMessageProcessor(
		u.command_topic,
		5*time.Minute,
		u.done,
		common.WrapThingCommandEntryProcessor(processCommands),
	)
}

func processCommands(m *common.ThingCommandEntry, err error) {
	log.Print("oh, look, a message!")
}
