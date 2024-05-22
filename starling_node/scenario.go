package starling_node

import (
	"bytes"
	"slices"
	"time"

	"github.com/starling-protocol/starling/device"
	"github.com/starling-protocol/starling/sync"
)

type EmptyScenario struct{}

func (e *EmptyScenario) OnStart(node *Node)                                                      {}
func (e *EmptyScenario) OnConnect(node *Node, address device.DeviceAddress)                      {}
func (e *EmptyScenario) OnDisconnect(node *Node, address device.DeviceAddress)                   {}
func (e *EmptyScenario) OnLog(node *Node, message string)                                        {}
func (e *EmptyScenario) OnReceivePacket(node *Node, packet []byte, address device.DeviceAddress) {}
func (e *EmptyScenario) OnReceiveData(node *Node, data []byte, session device.SessionID)         {}
func (e *EmptyScenario) OnSessionEstablished(node *Node, session device.SessionID, contact device.ContactID, address device.DeviceAddress) {
}
func (e *EmptyScenario) OnSessionBroken(node *Node, session device.SessionID)                  {}
func (e *EmptyScenario) OnProcessMessage(node *Node, session device.SessionID, message []byte) {}
func (e *EmptyScenario) OnMessageDelivered(node *Node, messageID device.MessageID)             {}
func (e *EmptyScenario) OnSyncStateChanged(node *Node, contact device.ContactID, stateUpdate sync.Model) {
}
func (e *EmptyScenario) OnTerminate(node *Node) {}

type Scenario interface {
	OnStart(node *Node)
	OnConnect(node *Node, address device.DeviceAddress)
	OnDisconnect(node *Node, address device.DeviceAddress)
	OnLog(node *Node, message string)
	OnReceivePacket(node *Node, packet []byte, address device.DeviceAddress)
	OnReceiveData(node *Node, data []byte, session device.SessionID)
	OnSessionEstablished(node *Node, session device.SessionID, contact device.ContactID, address device.DeviceAddress)
	OnSessionBroken(node *Node, session device.SessionID)
	OnProcessMessage(node *Node, session device.SessionID, message []byte)
	OnMessageDelivered(node *Node, messageID device.MessageID)
	OnSyncStateChanged(node *Node, contact device.ContactID, stateUpdate sync.Model)
	OnTerminate(node *Node)
}

type ScenarioBroadcastRouteRequest struct {
	EmptyScenario
	didBroadcast bool
	delay        time.Duration
}

func BroadcastRouteRequestScenario(delay time.Duration) *ScenarioBroadcastRouteRequest {
	return &ScenarioBroadcastRouteRequest{
		didBroadcast: false,
		delay:        delay,
	}
}

func (s *ScenarioBroadcastRouteRequest) OnStart(node *Node) {
	node.logf("scenario:broadcast_rreq:start:%d", node.ID())

	node.sim.DelayBy(func() {
		node.logf("scenario:broadcast_rreq:send:%d", node.ID())
		node.proto.BroadcastRouteRequest()
		s.didBroadcast = true
	}, s.delay)
}

func (s *ScenarioBroadcastRouteRequest) DidBroadcast() bool {
	return s.didBroadcast
}

type ScenarioTerminate struct {
	EmptyScenario
	didTerminate bool
}

func TerminateScenario() *ScenarioTerminate {
	return &ScenarioTerminate{
		didTerminate: false,
	}
}

func (s *ScenarioTerminate) OnStart(node *Node) {
	node.logf("scenario:terminate:terminated:%d", node.ID())
	s.didTerminate = true
	node.sim.Terminate(nil)
}

func (s *ScenarioTerminate) DidTerminate() bool {
	return s.didTerminate
}

type ScenarioAction struct {
	EmptyScenario
	didActivate bool
	action      func(node *Node)
}

func ActionScenario(action func(node *Node)) *ScenarioAction {
	return &ScenarioAction{
		didActivate: false,
		action:      action,
	}
}

func (s *ScenarioAction) OnStart(node *Node) {
	if !s.didActivate {
		node.logf("scenario:action:activated:%d", node.ID())
		s.didActivate = true
		s.action(node)
	}
}

type ScenarioMulti struct {
	EmptyScenario
	didActivate bool
	scenarios   []Scenario
}

func MultiScenario(scenarios ...Scenario) *ScenarioMulti {
	return &ScenarioMulti{
		didActivate: false,
		scenarios:   scenarios,
	}
}

func (s *ScenarioMulti) OnStart(node *Node) {
	if !s.didActivate {
		node.logf("scenario:multi:activated:%d", node.ID())
		s.didActivate = true
		for _, s := range s.scenarios {
			s.OnStart(node)
		}
	}
}

type ScenarioReply struct {
	EmptyScenario
	contact        device.ContactID
	receiveMessage string
	sendMessage    string
	sessions       []device.SessionID
	didDeliver     bool
	sendMessageID  *device.MessageID
}

func ReplyScenario(contact device.ContactID, receiveMessage string, sendMessage string) *ScenarioReply {
	return &ScenarioReply{
		contact:        contact,
		receiveMessage: receiveMessage,
		sendMessage:    sendMessage,
		sessions:       []device.SessionID{},
		didDeliver:     false,
		sendMessageID:  nil,
	}
}

func (s *ScenarioReply) DidSend() bool {
	return s.sendMessageID != nil
}

func (s *ScenarioReply) DidDeliver() bool {
	return s.didDeliver
}

func (s *ScenarioReply) OnSessionEstablished(node *Node, session device.SessionID, contact device.ContactID) {
	if contact == s.contact {
		s.sessions = append(s.sessions, session)
	}
}

func (s *ScenarioReply) OnReceiveData(node *Node, data []byte, session device.SessionID) {
	if slices.Contains(s.sessions, session) && bytes.Equal(data, []byte(s.receiveMessage)) {
		messageID, err := node.proto.SendMessage(session, []byte(s.sendMessage))
		if err != nil {
			node.logf("scenario:reply:error '%s'", err)
			return
		}

		s.sendMessageID = &messageID
	}
}

func (s *ScenarioReply) OnMessageDelivered(node *Node, messageID device.MessageID) {
	if s.sendMessageID != nil && *s.sendMessageID == messageID {
		s.didDeliver = true
	}
}

type ScenarioCount struct {
	EmptyScenario
	count    int
	scenario Scenario
}

func CountScenario(count int, scenario Scenario) *ScenarioCount {
	return &ScenarioCount{
		count:    count,
		scenario: scenario,
	}
}

func (s *ScenarioCount) Countdown() int {
	return s.count
}

func (s *ScenarioCount) OnStart(node *Node) {
	s.count -= 1

	if s.count == 0 && s.scenario != nil {
		s.scenario.OnStart(node)
	}
}

type ScenarioAll struct {
	triggers []*ScenarioAllEvent
	scenario Scenario
}

type ScenarioAllEvent struct {
	EmptyScenario
	parent    *ScenarioAll
	triggered bool
}

func AllEventsScenario() *ScenarioAll {
	return &ScenarioAll{
		triggers: []*ScenarioAllEvent{},
		scenario: nil,
	}
}

func (s *ScenarioAll) AllTriggered() bool {
	for _, trigger := range s.triggers {
		if !trigger.triggered {
			return false
		}
	}
	return true
}

func (s *ScenarioAll) NextEvent() *ScenarioAllEvent {
	trigger := ScenarioAllEvent{parent: s, triggered: false}
	s.triggers = append(s.triggers, &trigger)
	return &trigger
}

func (s *ScenarioAll) OnAllFinished(scenario Scenario) *ScenarioAll {
	s.scenario = scenario
	return s
}

func (s *ScenarioAllEvent) OnStart(node *Node) {
	s.triggered = true

	if s.parent.scenario != nil && s.parent.AllTriggered() {
		s.parent.scenario.OnStart(node)
	}
}

type ScenarioDelay struct {
	EmptyScenario
	scenario Scenario
	delay    time.Duration
}

func DelayScenario(scenario Scenario, delay time.Duration) *ScenarioDelay {
	return &ScenarioDelay{
		scenario: scenario,
		delay:    delay,
	}
}

func (s *ScenarioDelay) OnStart(node *Node) {
	node.sim.DelayBy(func() {
		s.delayDone(node)
	}, s.delay)
}

func (s *ScenarioDelay) delayDone(node *Node) {
	s.scenario.OnStart(node)
}
