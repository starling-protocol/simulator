package loggers

import (
	"fmt"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/pcap"
)

type PCAPLogger struct {
	pcap       *pcap.PCAPFile
	timeOffset time.Time
}

func NewPCAPLogger() *PCAPLogger {
	return &PCAPLogger{
		pcap: nil,
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (l *PCAPLogger) Init() {
	l.pcap = pcap.NewPCAPFile(time.Now())
	l.timeOffset = time.Now()
}

func (l *PCAPLogger) NewEvent(e simulator.Event) {

	switch e.EventType() {
	case simulator.RCV_MSG:

		receiveEvent := e.(*simulator.ReceiveEvent)

		if l.pcap == nil {
			panic("pcap_logger NewEvent called before Init")
		}
		l.pcap.AddPacket(receiveEvent.Packet(), l.timeOffset.Add(receiveEvent.Time()), fmt.Sprintf("%012x", int(receiveEvent.Target().NodeID())), fmt.Sprintf("%012x", int(receiveEvent.OriginNodeID())))
	// case simulator.SEND_MSG:

	// 	sendEvent := e.(*simulator.SendEvent)

	// 	if l.pcap == nil {
	// 		panic("pcap_logger NewEvent called before Init")
	// 	}
	// 	l.pcap.AddPacket(sendEvent.Packet(), l.timeOffset.Add(sendEvent.Time()), fmt.Sprintf("%012x", int(sendEvent.Origin().NodeID())), fmt.Sprintf("%012x", int(sendEvent.TargetNodeID())))

	case simulator.TERMINATE:
		l.pcap.WriteFile("capture")
	default:
		return
	}
}

func (l *PCAPLogger) Log(str string) {
}
