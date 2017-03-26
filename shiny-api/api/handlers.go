package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/apiarian/migration-playground/shiny-api/common"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type ThingView struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Foo       float64 `json:"foo"`
	CreatedOn string  `json:"created-on"`
	UpdatedOn string  `json:"updated-on"`
	Version   string  `json:"version"`
}

func ViewThing(t *common.Thing) *ThingView {
	if t == nil {
		return nil
	}

	return &ThingView{
		ID:        t.ID,
		Name:      t.Name,
		Foo:       t.Foo,
		CreatedOn: t.CreatedOn.Format(time.RFC3339),
		UpdatedOn: t.UpdatedOn.Format(time.RFC3339),
		Version:   t.Version,
	}
}

type ThingInput struct {
	Name    string  `json:"name"`
	Foo     float64 `json:"foo"`
	Version string  `json:"version"`
}

func CodeOrDefault(err error, def int) int {
	type coder interface {
		Code() int
	}

	c, ok := err.(coder)
	if ok {
		return c.Code()
	}

	return def
}

func MakeListThingsHandlerFunc(ts common.ThingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := ts.ListThings()
		if err != nil {
			WriteError(w, CodeOrDefault(err, http.StatusInternalServerError), err)
			return
		}

		WriteThings(w, t)
	}
}

func MakeGetThingHandlerFunc(ts common.ThingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id, ok := v["id"]
		if !ok {
			WriteError(w, http.StatusInternalServerError, errors.New("no id in request"))
			return
		}

		_, err := strconv.Atoi(id)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		t, err := ts.GetThing(id)
		if err != nil {
			WriteError(w, CodeOrDefault(err, http.StatusInternalServerError), err)
			return
		}

		WriteThing(w, t)
	}
}

func MakeCreateThingHandler(ts common.ThingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if err := r.Body.Close(); err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		var ti ThingInput
		if err := json.Unmarshal(body, &ti); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		t, err := ts.CreateThing(ti.Name, ti.Foo)
		if err != nil {
			WriteError(w, CodeOrDefault(err, http.StatusInternalServerError), err)
			return
		}

		WriteThing(w, t)
	}
}

func MakeUpdateThingHandlerFunc(ts common.ThingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id, ok := v["id"]
		if !ok {
			WriteError(w, http.StatusInternalServerError, errors.New("no id in request"))
			return
		}

		_, err := strconv.Atoi(id)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if err := r.Body.Close(); err != nil {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}

		var ti ThingInput
		if err := json.Unmarshal(body, &ti); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		t, err := ts.UpdateThing(id, ti.Version, ti.Name, ti.Foo)
		if err != nil {
			WriteError(w, CodeOrDefault(err, http.StatusInternalServerError), err)
			return
		}

		WriteThing(w, t)
	}
}

func MakeCheckCommandHandler(ts common.ThingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		cid, ok := v["id"]
		if !ok {
			WriteError(w, http.StatusInternalServerError, errors.New("no id in request"))
			return
		}

		t, err := ts.CheckCommand(cid)
		if err != nil {
			WriteError(w, CodeOrDefault(err, http.StatusInternalServerError), err)
			return
		}

		WriteThing(w, t)
	}
}

func WriteError(w http.ResponseWriter, c int, err error) {
	e := struct {
		Message string `json:"error-message"`
	}{
		Message: err.Error(),
	}

	w.WriteHeader(c)
	w.Header().Set("Content-Type", "application/json")

	ono := json.NewEncoder(w).Encode(&e)
	if ono != nil {
		panic(ono)
	}
}

func WriteThing(w http.ResponseWriter, t *common.Thing) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(ViewThing(t))
	if err != nil {
		panic(err)
	}
}

func WriteThings(w http.ResponseWriter, ts []*common.Thing) {
	tvs := make([]*ThingView, len(ts))
	for i, t := range ts {
		tvs[i] = ViewThing(t)
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(&tvs)
	if err != nil {
		panic(err)
	}
}
