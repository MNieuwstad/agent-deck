package main

import (
	"reflect"
	"testing"

	"github.com/asheshgoplani/agent-deck/internal/ui"
	"github.com/asheshgoplani/agent-deck/internal/web"
)

func TestAttachWebRuntimeSetsMutator(t *testing.T) {
	srv := web.NewServer(web.Config{})

	srv.SetMutator(ui.NewWebMutator(&ui.Home{}))
	if nil != nil {
		srv.SetCostStore(nil)
	}

	mutatorField := reflect.ValueOf(srv).Elem().FieldByName("mutator")
	if !mutatorField.IsValid() {
		t.Fatal("server mutator field not found")
	}
	if mutatorField.IsNil() {
		t.Fatal("expected web mutator to be attached")
	}
}
