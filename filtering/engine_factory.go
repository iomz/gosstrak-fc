// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package filtering

import (
	"log"
	"reflect"
	"unsafe"
)

// ManagementMessageType is to indicate the type of ManagementMessage
type ManagementMessageType int

// ManagementMessage types
const (
	AddSubscription ManagementMessageType = iota
	DeleteSubscription
	OnEngineGenerated
	DeployEngine
)

// ManagementMessage holds management action for the EngineFactory
type ManagementMessage struct {
	Type                    ManagementMessageType
	FilterString            string
	NotificationURI         string
	EngineGeneratorInstance *EngineGenerator
}

// EngineFactory manages the FC's subscriptions and engine instances
type EngineFactory struct {
	mainChannel          chan ManagementMessage
	generatorChannels    []chan ManagementMessage
	currentSubscriptions Subscriptions
	productionLines      []*EngineGenerator
	deploymentPriority   map[string]uint8
	currentEngine        Engine
}

// IsActive returns false if no engine is available
func (ef *EngineFactory) IsActive() bool {
	if ef.currentEngine == nil {
		return false
	}
	return true
}

// Search is a wrapper for Search() with the currentEngine
func (ef *EngineFactory) Search(id []byte) []string {
	return ef.currentEngine.Search(id)
}

// NewEngineFactory returns the pointer to a new EngineFactory instance
func NewEngineFactory(sub Subscriptions, mc chan ManagementMessage) *EngineFactory {
	ef := &EngineFactory{
		mainChannel: mc,
	}

	// Load saved subscriptions?
	ef.currentSubscriptions = sub

	// Load all the possible engines
	ef.productionLines = []*EngineGenerator{}
	ef.generatorChannels = []chan ManagementMessage{}
	for name, constructor := range AvailableEngines {
		ch := make(chan ManagementMessage)
		ef.generatorChannels = append(ef.generatorChannels, ch)
		eg := NewEngineGenerator(name, constructor, ch)
		ef.productionLines = append(ef.productionLines, eg)
	}

	// Calculate the priority of deployment
	ef.deploymentPriority = map[string]uint8{}
	priority := uint8(0)
	for name := range AvailableEngines {
		ef.deploymentPriority[name] = priority
		priority++
	}

	log.Printf("[EngineFactory] deploymentPriority: %v", ef.deploymentPriority)

	return ef
}

// Run starts the engine factory to react with the ManagementChannel
func (ef *EngineFactory) Run() {
	log.Println("[EngineFactory] start running")
	// set channels from EngineGenerators + main
	cases := make([]reflect.SelectCase, len(ef.generatorChannels)+1)
	for i, ch := range append(ef.generatorChannels, ef.mainChannel) {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	go func() {
		log.Println("[EngineFactory] setting up managementChannel listener")
		for {
			_, val, ok := reflect.Select(cases)
			if !ok {
				break
			}
			//msg, _ := reflect.ValueOf(val).Interface().(ManagementMessage)
			msg := ManagementMessage{
				Type:                    val.FieldByName("Type").Interface().(ManagementMessageType),
				FilterString:            val.FieldByName("FilterString").String(),
				NotificationURI:         val.FieldByName("FilterString").String(),
				EngineGeneratorInstance: (*EngineGenerator)(unsafe.Pointer(val.FieldByName("EngineGeneratorInstance").Pointer())),
			}
			switch msg.Type {
			case AddSubscription:
				if _, ok := ef.currentSubscriptions[msg.FilterString]; !ok {
					ef.currentSubscriptions[msg.FilterString] = &Info{
						Offset:          0,
						NotificationURI: msg.NotificationURI,
					}
					for _, eg := range ef.productionLines {
						err := eg.FSM.Event("update", &msg)
						if err != nil {
							log.Println(err)
						}
					}
				}
			case DeleteSubscription:
				if _, ok := ef.currentSubscriptions[msg.FilterString]; ok {
					delete(ef.currentSubscriptions, msg.FilterString)
					for _, eg := range ef.productionLines {
						err := eg.FSM.Event("update", &msg)
						if err != nil {
							log.Println(err)
						}
					}
				}
			case OnEngineGenerated:
				log.Printf("[EngineFactory] received OnEngineGenerated from %s", msg.EngineGeneratorInstance.Engine.Name())
				if ef.currentEngine == nil {
					log.Printf("[EngineFactory] set %s as an initial engine", msg.EngineGeneratorInstance.Name)
					ef.currentEngine = msg.EngineGeneratorInstance.Engine
					continue
				}
				if ef.deploymentPriority[ef.currentEngine.Name()] < ef.deploymentPriority[msg.EngineGeneratorInstance.Engine.Name()] {
					log.Printf("[EngineFactory] %s replaces the currentEngine %s", msg.EngineGeneratorInstance.Name, ef.currentEngine.Name())
					ef.currentEngine = msg.EngineGeneratorInstance.Engine
					continue
				}
				log.Printf("[EngineFactory] %s didn't replace the currentEngine %s", msg.EngineGeneratorInstance.Name, ef.currentEngine.Name())
			}
		}
		log.Fatalln("mainChannel listener exited in gosstrak-fc")
	}()

	// initialize the engines
	log.Println("[EngineFactory] initializing engines")
	for _, eg := range ef.productionLines {
		// pass the cloned subscriptions
		eg.FSM.Event("init", ef.currentSubscriptions.Clone())
	}
}
