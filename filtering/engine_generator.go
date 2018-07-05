// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package filtering

import (
	"log"
	//"reflect"

	"github.com/looplab/fsm"
)

// EngineGenerator produce an engine according to the FSM
type EngineGenerator struct {
	managementChannel chan ManagementMessage
	Name              string
	Engine            Engine
	FSM               *fsm.FSM
}

// NewEngineGenerator returns the pointer to a new EngineGenerator instance
func NewEngineGenerator(name string, ec EngineConstructor, mc chan ManagementMessage) *EngineGenerator {
	eg := &EngineGenerator{
		managementChannel: mc,
		Name:              name,
	}

	eg.FSM = fsm.NewFSM(
		"unavailable",
		fsm.Events{
			{Name: "init", Src: []string{"unavailable"}, Dst: "generating"},
			{Name: "deploy", Src: []string{"generating", "rebuilding"}, Dst: "ready"},
			{Name: "update", Src: []string{"ready"}, Dst: "pending"},
			{Name: "rebuild", Src: []string{"pending"}, Dst: "rebuilding"},
		},
		fsm.Callbacks{
			"enter_state":      func(e *fsm.Event) { eg.enterState(e) },
			"enter_generating": func(e *fsm.Event) { eg.enterGenerating(e) },
			"enter_ready":      func(e *fsm.Event) { eg.enterReady(e) },
			"enter_pending":    func(e *fsm.Event) { eg.enterPending(e) },
			"enter_rebuilding": func(e *fsm.Event) { eg.enterRebuilding(e) },
		},
	)

	return eg
}

func (eg *EngineGenerator) enterState(e *fsm.Event) {
	log.Printf("[EngineGenerator] %s event, %s entering %s", e.Event, eg.Name, e.Dst)
}

func (eg *EngineGenerator) enterGenerating(e *fsm.Event) {
	go func() {
		//log.Printf("[EngineGenerator] start generating %s engine", eg.Name)
		sub := e.Args[0].(Subscriptions)
		eg.Engine = AvailableEngines[eg.Name](sub)
		eg.FSM.Event("deploy")
	}()
}

func (eg *EngineGenerator) enterRebuilding(e *fsm.Event) {
	msg := e.Args[0].(*ManagementMessage)
	switch msg.Type {
	case AddSubscription:
		eg.Engine.AddSubscription(Subscriptions{
			msg.FilterString: &Info{
				Offset:          0,
				NotificationURI: msg.NotificationURI,
			},
		})
	case DeleteSubscription:
		eg.Engine.DeleteSubscription(Subscriptions{
			msg.FilterString: &Info{
				Offset:          0,
				NotificationURI: msg.NotificationURI,
			},
		})
	}
	eg.FSM.Event("deploy")
}

func (eg *EngineGenerator) enterReady(e *fsm.Event) {
	log.Printf("[EngineGenerator] finished gererating %s engine", eg.Name)
	eg.managementChannel <- ManagementMessage{
		Type: OnEngineGenerated,
		EngineGeneratorInstance: eg,
	}
}

func (eg *EngineGenerator) enterPending(e *fsm.Event) {
	// Wait until the engine finishes the current execution
	eg.FSM.Event("rebuild", e.Args[0].(*ManagementMessage))
}
