package loggers

import (
	"fmt"

	"github.com/starling-protocol/simulator"
)

type StandardLogger struct {
	lastEvent simulator.Event
}

func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		lastEvent: nil,
	}
}

func (l *StandardLogger) Init() {
	fmt.Print("Starting logging")
}

var default_color = "\033[39m"
var cyan = "\033[36m"
var blue = "\033[94m"
var yellow = "\x1b[33m"

func (l *StandardLogger) NewEvent(e simulator.Event) {
	switch e.EventType() {
	case simulator.TIMESTEP:
		if l.lastEvent != nil && l.lastEvent.EventType() == simulator.TIMESTEP {
			break
		}
		fmt.Printf(yellow+"[EVENT] %d %d: [STEP]\n", e.Time().Milliseconds(), e.SequenceNumber())
	case simulator.CONNECT:
		connectEvent := e.(*simulator.ConnectEvent)
		fmt.Printf(cyan+"[EVENT] %d %d: [CONNECT] %d, %d\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), connectEvent.NodeA().NodeID(), connectEvent.NodeB().NodeID())
	case simulator.DISCONNECT:
		disconnectEvent := e.(*simulator.DisconnectEvent)
		fmt.Printf(cyan+"[EVENT] %d %d: [DISCONNECT] %d, %d\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), disconnectEvent.NodeA().NodeID(), disconnectEvent.NodeB().NodeID())
	case simulator.ADD_NODE:
		addNodeEvent := e.(*simulator.AddNodeEvent)
		fmt.Printf(cyan+"[EVENT] %d %d: [ADD_NODE] %d\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), addNodeEvent.Node().NodeID())
	case simulator.SEND_MSG:
		sendEvent := e.(*simulator.SendEvent)
		fmt.Printf(blue+"[EVENT] %d %d: [SEND] %d to %d\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), sendEvent.Origin().NodeID(), sendEvent.TargetNodeID())
	case simulator.RCV_MSG:
		receiveEvent := e.(*simulator.ReceiveEvent)
		fmt.Printf(blue+"[EVENT] %d %d: [RECEIVE] %d from %d\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), receiveEvent.Target().NodeID(), receiveEvent.OriginNodeID())
	case simulator.DELAY:
		fmt.Printf(blue+"[EVENT] %d %d: [DELAY] \n"+default_color, e.Time().Milliseconds(), e.SequenceNumber())
	case simulator.TERMINATE:
		terminateEvent := e.(*simulator.TerminateEvent)
		if terminateEvent.Error() == nil {
			fmt.Printf(blue+"[EVENT] %d %d: [TERMINATE]\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber())
		} else {
			fmt.Printf(blue+"[EVENT] %d %d: [TERMINATE] with error: '%s'\n"+default_color, e.Time().Milliseconds(), e.SequenceNumber(), terminateEvent.Error())
		}
	default:
		panic("Simulator event error!")
	}

	l.lastEvent = e
}

func (l *StandardLogger) Log(str string) {
	fmt.Printf("\t[DEBUG] %d %d: %s\n", l.lastEvent.Time().Milliseconds(), l.lastEvent.SequenceNumber(), str)
}
