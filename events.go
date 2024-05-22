package simulator

import (
	"container/heap"
	"time"
)

type EventType int64

const (
	TIMESTEP EventType = iota
	CONNECT
	DISCONNECT
	ADD_NODE
	RM_NODE
	SEND_MSG
	RCV_MSG
	DELAY
	TERMINATE
)

func (e EventType) String() string {
	switch e {
	case TIMESTEP:
		return "timestep"
	case CONNECT:
		return "connect"
	case DISCONNECT:
		return "disconnect"
	case RCV_MSG:
		return "rcv_msg"
	case SEND_MSG:
		return "send_msg"
	case DELAY:
		return "delay"
	case TERMINATE:
		return "terminate"
	default:
		return "unknown"
	}
}

type BaseEvent struct {
	time           time.Duration
	sequenceNumber int64
	parentEvent    Event
}

func (e *BaseEvent) Time() time.Duration {
	return e.time
}

func (e *BaseEvent) SequenceNumber() int64 {
	return e.sequenceNumber
}

func (e *BaseEvent) ParentEvent() Event {
	return e.parentEvent
}

type Event interface {
	EventType() EventType
	Time() time.Duration
	SequenceNumber() int64
	ParentEvent() Event
}

func NodeFromEvent(e Event) Node {
	switch e.EventType() {
	case CONNECT:
		conn := e.(*ConnectEvent)
		return conn.nodeA.node
	case DISCONNECT:
		conn := e.(*DisconnectEvent)
		return conn.nodeA.node
	case SEND_MSG:
		send := e.(*SendEvent)
		return send.origin.node
	case RCV_MSG:
		rcv := e.(*ReceiveEvent)
		return rcv.target.node
	case DELAY:
		delay := e.(*DelayEvent)
		return delay.node.node
	case TERMINATE:
		return NodeFromEvent(e.ParentEvent())
	default:
		return nil
	}
}

type TimestepEvent struct {
	BaseEvent
}

func (e *TimestepEvent) EventType() EventType {
	return TIMESTEP
}

type DelayEvent struct {
	BaseEvent
	node           *InternalNode
	functionToCall func()
}

func (e *DelayEvent) EventType() EventType {
	return DELAY
}

type ConnectEvent struct {
	BaseEvent
	nodeA *InternalNode
	nodeB *InternalNode
}

func (s *Simulator) NewConnectEvent(time time.Duration, nodeA *InternalNode, nodeB *InternalNode) *ConnectEvent {
	return &ConnectEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		nodeA: nodeA,
		nodeB: nodeB,
	}
}

func (e *ConnectEvent) EventType() EventType {
	return CONNECT
}

func (e *ConnectEvent) NodeA() *InternalNode {
	return e.nodeA
}

func (e *ConnectEvent) NodeB() *InternalNode {
	return e.nodeB
}

type DisconnectEvent struct {
	BaseEvent
	nodeA *InternalNode
	nodeB *InternalNode
}

func (s *Simulator) NewDisconnectEvent(time time.Duration, nodeA *InternalNode, nodeB *InternalNode) *DisconnectEvent {
	return &DisconnectEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		nodeA: nodeA,
		nodeB: nodeB,
	}
}

func (e *DisconnectEvent) EventType() EventType {
	return DISCONNECT
}

func (e *DisconnectEvent) NodeA() *InternalNode {
	return e.nodeA
}

func (e *DisconnectEvent) NodeB() *InternalNode {
	return e.nodeB
}

type AddNodeEvent struct {
	BaseEvent
	node *InternalNode
}

func (e *AddNodeEvent) EventType() EventType {
	return ADD_NODE
}

func (e *AddNodeEvent) Node() *InternalNode {
	return e.node
}

type RemoveNodeEvent struct {
	BaseEvent
	node *InternalNode
}

func (e *RemoveNodeEvent) EventType() EventType {
	return RM_NODE
}

func (e *RemoveNodeEvent) Node() *InternalNode {
	return e.node
}

type SendEvent struct {
	BaseEvent
	target          *InternalNode
	origin          *InternalNode
	targetNodeID    NodeID
	peer            Peer
	shouldBeDropped bool
	delay           time.Duration
	packet          []byte
}

func (e *SendEvent) EventType() EventType {
	return SEND_MSG
}

func (e *SendEvent) Packet() []byte {
	return e.packet
}

func (e *SendEvent) TargetNodeID() NodeID {
	return e.targetNodeID
}

func (e *SendEvent) Origin() *InternalNode {
	return e.origin
}

type ReceiveEvent struct {
	BaseEvent
	target       *InternalNode
	origin       *InternalNode
	originNodeID NodeID
	peer         Peer
	packet       []byte
}

func (e *ReceiveEvent) EventType() EventType {
	return RCV_MSG
}

func (e *ReceiveEvent) Packet() []byte {
	return e.packet
}

func (e *ReceiveEvent) OriginNodeID() NodeID {
	return e.originNodeID
}

func (e *ReceiveEvent) Target() *InternalNode {
	return e.target
}

type TerminateEvent struct {
	BaseEvent
	err error
}

func (e *TerminateEvent) EventType() EventType {
	return TERMINATE
}

func (e *TerminateEvent) Error() error {
	return e.err
}

func (s *Simulator) pushTimeStepEvent(time time.Duration) {
	event := &TimestepEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushConnectEvent(time time.Duration, nodeA *InternalNode, nodeB *InternalNode) {
	event := &ConnectEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		nodeA: nodeA,
		nodeB: nodeB,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushDisconnectEvent(time time.Duration, nodeA *InternalNode, nodeB *InternalNode) {
	event := &DisconnectEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		nodeA: nodeA,
		nodeB: nodeB,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushAddNodeEvent(time time.Duration, node *InternalNode) {
	event := &AddNodeEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		node: node,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushRemoveNodeEvent(time time.Duration, node *InternalNode) {
	event := &RemoveNodeEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		node: node,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushReceiveEvent(time time.Duration, target *InternalNode, origin *InternalNode, originNodeID NodeID, peer Peer, packet []byte) {
	event := &ReceiveEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		target:       target,
		origin:       origin,
		originNodeID: originNodeID,
		peer:         peer,
		packet:       packet,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushSendEvent(time time.Duration, target *InternalNode, origin *InternalNode, targetNodeID NodeID, peer Peer, shouldBeDropped bool, delay time.Duration, packet []byte) {
	event := &SendEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		target:          target,
		origin:          origin,
		targetNodeID:    targetNodeID,
		peer:            peer,
		shouldBeDropped: shouldBeDropped,
		delay:           delay,
		packet:          packet,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushDelayEvent(time time.Duration, node *InternalNode, functionToCall func()) {
	event := &DelayEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: s.currentSequenceNumber,
			parentEvent:    s.lastEvent,
		},
		node:           node,
		functionToCall: functionToCall,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}

func (s *Simulator) pushTerminateEvent(time time.Duration, err error) {
	event := &TerminateEvent{
		BaseEvent: BaseEvent{
			time:           time,
			sequenceNumber: -1,
			parentEvent:    s.lastEvent,
		},
		err: err,
	}
	heap.Push(s.eventQueue, event)
	s.currentSequenceNumber++
}
