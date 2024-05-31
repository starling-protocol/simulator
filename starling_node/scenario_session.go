package starling_node

import (
	"time"

	"github.com/starling-protocol/starling/device"
)

type SessionOrchestrator struct {
	contact   device.ContactID
	sessionID *device.SessionID
	count     int
}

func NewSessionOrchestrator(contact device.ContactID) *SessionOrchestrator {
	return &SessionOrchestrator{
		contact:   contact,
		sessionID: nil,
		count:     0,
	}
}

type ScenarioSession struct {
	EmptyScenario
	orchestrator    *SessionOrchestrator
	sessionID       *device.SessionID
	maintainSession bool
	cancelRetry     *bool
	didEstablish    bool
	didBreak        bool
	// synchronizeScenarios []*ScenarioSynchronize
	sendDataScenarios []*ScenarioSendData
}

func (s *SessionOrchestrator) SessionScenario(maintainSession bool) *ScenarioSession {
	return &ScenarioSession{
		orchestrator:    s,
		maintainSession: maintainSession,
		cancelRetry:     nil,
		didBreak:        false,
	}
}

func (s *ScenarioSession) CancelRetry() {
	*s.cancelRetry = true
}

func (s *ScenarioSession) DidEstablishSession() bool {
	return s.didEstablish
}

func (s *ScenarioSession) DidBreakSession() bool {
	return s.didBreak
}

func (s *ScenarioSession) HasSession() bool {
	return s.sessionID != nil
}

type receiveDataArgs struct {
	session *ScenarioSession
	message string
}

func (s *ScenarioSession) ReceiveDataEvent(message string) Event {
	return Event{
		eventType: eventReceiveData,
		args:      receiveDataArgs{session: s, message: message},
	}
}

// type ScenarioSynchronize struct {
// 	EmptyScenario
// 	context *ScenarioSession
// 	didSend bool
// }

// func (s *ScenarioSynchronize) OnStart(node *Node) {
// 	if s.context.sessionID != nil {
// 		node.sim.DelayBy(func() {
// 			node.proto.BroadcastRouteRequest()
// 			s.didSend = true
// 		}, 20*time.Millisecond)
// 	} else {
// 		if !slices.Contains(s.context.synchronizeScenarios, s) {
// 			s.context.synchronizeScenarios = append(s.context.synchronizeScenarios, s)
// 		}
// 	}
// }

func (s *ScenarioSession) OnStart(node *Node) {
	s.establishConnection(node)
}

func (s *ScenarioSession) establishConnection(node *Node) {
	node.sim.DelayBy(func() {
		node.proto.BroadcastRouteRequest()
	}, 20*time.Millisecond)

	if s.maintainSession {
		if s.cancelRetry != nil {
			*s.cancelRetry = true
		}

		cancelRetry := false
		s.cancelRetry = &cancelRetry

		node.sim.DelayBy(func() {
			if s.sessionID == nil && !cancelRetry {
				s.establishConnection(node)
			}
		}, 60*time.Second)
	}
}

func (s *ScenarioSession) sessionEstablished(node *Node, sessionID device.SessionID) {
	for _, sendDataScenario := range s.sendDataScenarios {
		sendDataScenario.sendData(node)
	}
}

func (s *ScenarioSession) OnSessionEstablished(node *Node, session device.SessionID, contact device.ContactID, address device.DeviceAddress) {
	if s.sessionID == nil {
		if s.orchestrator.contact == contact && (s.orchestrator.sessionID == nil || *s.orchestrator.sessionID == session) {
			s.orchestrator.sessionID = &session
			s.orchestrator.count += 1
			s.sessionID = &session
			s.didEstablish = true

			// if s.cancelRetry != nil {
			// 	*s.cancelRetry = true
			// }

			s.sessionEstablished(node, *s.orchestrator.sessionID)
		}
	}
}

func (s *ScenarioSession) OnSessionBroken(node *Node, session device.SessionID) {
	if s.sessionID != nil && *s.sessionID == session {
		s.orchestrator.count -= 1
		if s.orchestrator.count <= 0 {
			s.orchestrator.sessionID = nil
		}

		s.didBreak = true
		s.sessionID = nil

		if s.maintainSession {
			s.establishConnection(node)
		}
	}
}

type ScenarioSendData struct {
	EmptyScenario
	context          *ScenarioSession
	data             string
	didSendOnSession *device.SessionID
	didDeliver       bool
	eventArgs        *messageDeliveredArgs
}

func (s *ScenarioSession) SendDataScenario(data string) *ScenarioSendData {
	sendDataScenario := &ScenarioSendData{
		context:          s,
		data:             data,
		didSendOnSession: nil,
		didDeliver:       false,
		eventArgs: &messageDeliveredArgs{
			messageID: 0,
		},
	}

	sendDataScenario.eventArgs.didDeliver = func() {
		sendDataScenario.didDeliver = true
	}

	return sendDataScenario
}

func (s *ScenarioSendData) OnStart(node *Node) {
	node.logf("scenario:send_data:start:%d '%s'", node.ID(), s.data)
	s.context.sendDataScenarios = append(s.context.sendDataScenarios, s)
	s.sendData(node)
}

func (s *ScenarioSendData) sendData(node *Node) {
	if s.context.sessionID != nil && !s.didDeliver && (s.didSendOnSession == nil || *s.didSendOnSession != *s.context.sessionID) {
		sessionID := *s.context.sessionID

		node.logf("scenario:send_data:%d '%s'", sessionID, s.data)
		msgID, err := node.proto.SendMessage(sessionID, []byte(s.data))
		if err != nil {
			node.logf("scenario:send_data:error '%s'", err)
			return
		}

		s.eventArgs.messageID = msgID
		s.didSendOnSession = &sessionID
	}
}

func (s *ScenarioSendData) DidSend() bool {
	return s.didSendOnSession != nil
}

func (s *ScenarioSendData) DeliveredEvent() Event {
	return Event{
		eventType: eventMessageDelivered,
		args:      s.eventArgs,
	}
}
