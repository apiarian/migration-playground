package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
)

type Updater struct {
	kc            *common.KafkaClient
	command_topic string
	done          chan struct{}
	assignIDs     bool
	nextID        int
	mux           *sync.Mutex
	entityCache   map[string]*common.Thing
}

func NewUpdater(
	kc *common.KafkaClient,
	command_topic string,
	assignIDs bool,
	nextID int,
) *Updater {
	return &Updater{
		kc:            kc,
		command_topic: command_topic,
		done:          make(chan struct{}),
		assignIDs:     assignIDs,
		nextID:        nextID,
		mux:           &sync.Mutex{},
		entityCache:   make(map[string]*common.Thing),
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
		common.WrapThingCommandEntryProcessor(u.MakeCommandProcessor()),
	)
}

func (u *Updater) MakeCommandProcessor() func(*common.ThingCommandEntry, error) {
	return func(m *common.ThingCommandEntry, err error) {
		if err != nil {
			log.Printf("error before command could be processed: %v", err)
			return
		}

		if m.Request != nil {
			switch m.Type {
			case common.CreateCommandType:
				var id string
				if m.Request.ID == "" {
					if u.assignIDs {
						u.mux.Lock()
						i := u.nextID
						u.nextID = u.nextID + 1
						id = strconv.Itoa(i)
						u.mux.Unlock()
					} else {
						log.Printf("cannot assign IDs; ignoring request")
						return
					}
				} else {
					id = m.Request.ID
				}

				t := &common.Thing{
					ID:        id,
					Name:      m.Request.Name,
					Foo:       m.Request.Foo,
					CreatedOn: m.Request.CreatedOn,
					UpdatedOn: m.Request.UpdatedOn,
					Version:   m.Request.Version,
				}

				u.mux.Lock()
				defer u.mux.Unlock()
				_, exists := u.entityCache[id]
				if exists {
					err := u.kc.PublishThingErrorCommand(
						m.CID,
						http.StatusConflict,
						"encountered id conflict",
						m.Type,
					)
					if err != nil {
						log.Printf("error publishing thing error message: %v", err)
					}
					return
				}
				err := u.kc.PublishThingSuccessCommand(m.CID, t, m.Type)
				if err != nil {
					log.Printf("error publishing thing success message: %v", err)
					return
				}
				u.entityCache[id] = t

			case common.UpdateCommandType:
				log.Printf("update commands are not supported yet")

			default:
				log.Printf("unknown command request type: %s", m.Type)
			}
		}
	}
}
