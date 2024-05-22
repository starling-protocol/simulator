package starling_node

import (
	"bytes"
	"strings"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/starling/device"
	"github.com/starling-protocol/starling/sync"
)

type EventType uint

const (
	eventReceiveData EventType = iota
	eventSessionEstablished
	eventConnection
	eventDisconnection
	eventLog
	eventProcessMessage
	eventMessageDelivered
	eventSyncStateUpdated
	eventTerminate
)

type Event struct {
	eventType EventType
	args      interface{}
}

type ScenarioEvent struct {
	event      Event
	scenarios  []Scenario
	hasStarted bool
}

type connectArgs struct {
	nodeID simulator.NodeID
}

func ConnectEvent(nodeID simulator.NodeID) Event {
	return Event{
		eventType: eventConnection,
		args:      connectArgs{nodeID: nodeID},
	}
}

type disconnectArgs struct {
	nodeID simulator.NodeID
}

func DisconnectEvent(nodeID simulator.NodeID) Event {
	return Event{
		eventType: eventDisconnection,
		args:      disconnectArgs{nodeID: nodeID},
	}
}

type logArgs struct {
	prefix string
}

func LogEvent(prefix string) Event {
	return Event{
		eventType: eventLog,
		args:      logArgs{prefix: prefix},
	}
}

type sessionEstablishedArgs struct {
	contact device.ContactID
}

func SessionEstablishedEvent(contact device.ContactID) Event {
	return Event{
		eventType: eventSessionEstablished,
		args: sessionEstablishedArgs{
			contact: contact,
		},
	}
}

type processMessageArgs struct {
	messageID device.MessageID
	session   device.SessionID
	message   []byte
}

func processMessageEvent(session device.SessionID, message []byte) Event {
	return Event{
		eventType: eventProcessMessage,
		args: processMessageArgs{
			session: session,
			message: message,
		},
	}
}

type messageDeliveredArgs struct {
	messageID  device.MessageID
	didDeliver func()
}

func MessageDeliveredEvent(messageID device.MessageID) Event {
	return Event{
		eventType: eventMessageDelivered,
		args:      messageDeliveredArgs{messageID: messageID},
	}
}

type syncStateChangedArgs struct {
	contact device.ContactID
	filter  func(sync.Model) bool
}

func SyncStateChangedEvent(contact device.ContactID, filter func(sync.Model) bool) Event {
	return Event{
		eventType: eventSyncStateUpdated,
		args: syncStateChangedArgs{
			contact: contact,
			filter:  filter,
		},
	}
}

func TerminateEvent() Event {
	return Event{
		eventType: eventTerminate,
		args:      nil,
	}
}

func EventScenario(event Event) *ScenarioEvent {
	return &ScenarioEvent{
		event:      event,
		scenarios:  []Scenario{},
		hasStarted: false,
	}
}

func (s *ScenarioEvent) DidTrigger() bool {
	return s.hasStarted
}

func (s *ScenarioEvent) OnEvent(scenario Scenario) *ScenarioEvent {
	if scenario == nil {
		return s
	}

	s.scenarios = append(s.scenarios, scenario)
	return s
}

func (s *ScenarioEvent) OnStart(node *Node) {}

func (s *ScenarioEvent) OnConnect(node *Node, address device.DeviceAddress) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnConnect(node, address)
		}
	} else {
		if s.event.eventType == eventConnection {
			eventArgs := s.event.args.(connectArgs)
			if idToAddr(eventArgs.nodeID) == address {
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnDisconnect(node *Node, address device.DeviceAddress) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnDisconnect(node, address)
		}
	} else {
		if s.event.eventType == eventDisconnection {
			eventArgs := s.event.args.(disconnectArgs)
			if idToAddr(eventArgs.nodeID) == address {
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnLog(node *Node, message string) {
	if s.event.eventType == eventLog {
		eventArgs := s.event.args.(logArgs)
		if strings.HasPrefix(message, eventArgs.prefix) {
			for _, scenario := range s.scenarios {
				scenario.OnLog(node, message)
			}
		}
	}
}

func (s *ScenarioEvent) OnReceivePacket(node *Node, packet []byte, address device.DeviceAddress) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnReceivePacket(node, packet, address)
		}
	}
}

func (s *ScenarioEvent) OnReceiveData(node *Node, data []byte, session device.SessionID) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnReceiveData(node, data, session)
		}
	} else {
		if s.event.eventType == eventReceiveData {
			eventArgs := s.event.args.(receiveDataArgs)

			if eventArgs.session.sessionID != nil && *eventArgs.session.sessionID == session && bytes.Equal([]byte(eventArgs.message), data) {
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnSessionEstablished(node *Node, session device.SessionID, contact device.ContactID, address device.DeviceAddress) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnSessionEstablished(node, session, contact, address)
		}
	} else {
		if s.event.eventType == eventSessionEstablished {
			eventArgs := s.event.args.(sessionEstablishedArgs)
			if eventArgs.contact == contact {
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnSessionBroken(node *Node, session device.SessionID) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnSessionBroken(node, session)
		}
	}
}

func (s *ScenarioEvent) OnProcessMessage(node *Node, session device.SessionID, message []byte) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnProcessMessage(node, session, message)
		}
	} else {
		if s.event.eventType == eventProcessMessage {
			eventArgs := s.event.args.(*processMessageArgs)

			if eventArgs.session == session && bytes.Equal(eventArgs.message, message) {
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnMessageDelivered(node *Node, messageID device.MessageID) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnMessageDelivered(node, messageID)
		}
	} else {
		if s.event.eventType == eventMessageDelivered {
			eventArgs := s.event.args.(*messageDeliveredArgs)

			if eventArgs.messageID == messageID {
				eventArgs.didDeliver()
				s.eventTriggered(node)
			}
		}
	}
}

func (s *ScenarioEvent) OnSyncStateChanged(node *Node, contact device.ContactID, stateUpdate sync.Model) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnSyncStateChanged(node, contact, stateUpdate)
		}
	} else {
		if s.event.eventType == eventSyncStateUpdated {
			eventArgs := s.event.args.(syncStateChangedArgs)

			if eventArgs.contact == contact && eventArgs.filter(stateUpdate) {
				s.eventTriggered(node)
				for _, scenario := range s.scenarios {
					scenario.OnSyncStateChanged(node, contact, stateUpdate)
				}
			}
		}
	}
}

func (s *ScenarioEvent) OnTerminate(node *Node) {
	if s.hasStarted {
		for _, scenario := range s.scenarios {
			scenario.OnTerminate(node)
		}
	} else {
		if s.event.eventType == eventTerminate {
			s.eventTriggered(node)
			for _, scenario := range s.scenarios {
				scenario.OnTerminate(node)
			}
		}
	}
}

func (s *ScenarioEvent) eventTriggered(node *Node) {
	node.logf("scenario:event:triggered:%d", node.ID())

	s.hasStarted = true
	for _, scenario := range s.scenarios {
		scenario.OnStart(node)
	}
}
