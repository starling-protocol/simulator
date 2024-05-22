package loggers

import (
	"fmt"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/starling/network_layer"
	"github.com/starling-protocol/starling/packet_layer"
)

type StatisticsLogger struct {
	realStartTime time.Time

	packagesSent     int64
	packagesReceived int64
	bytesReceived    int64
	metaDataPackets  int64
	metadataBytes    int64

	networkLevelPackages int64
	routeRequests        int64
	routeReplies         int64
	routeErrors          int64
	routeSessionData     int64
}

func NewStatisticsLogger() *StatisticsLogger {
	return &StatisticsLogger{
		packagesSent:         0,
		packagesReceived:     0,
		bytesReceived:        0,
		metaDataPackets:      0,
		metadataBytes:        0,
		networkLevelPackages: 0,
		routeRequests:        0,
		routeReplies:         0,
		routeErrors:          0,
		routeSessionData:     0,
	}
}

func (l *StatisticsLogger) Init() {
	l.realStartTime = time.Now()
}

func (l *StatisticsLogger) NewEvent(e simulator.Event) {

	switch e.EventType() {
	case simulator.RCV_MSG:
		receiveEvent := e.(*simulator.ReceiveEvent)
		l.packagesReceived++
		l.bytesReceived += int64(len(receiveEvent.Packet()))

		decoder := packet_layer.NewPacketDecoder()
		decoder.AppendPacket(receiveEvent.Packet())
		decodedMsg, err := decoder.ReadMessage()
		if err == nil {
			networkPacket, err := network_layer.DecodeRoutingPacket(decodedMsg)
			if err == nil {
				l.networkLevelPackages++
				if networkPacket.PacketType() != network_layer.SESS {
					l.metaDataPackets++
					l.metadataBytes += int64(len(decodedMsg))
				}

				switch networkPacket.PacketType() {
				case network_layer.RREQ:
					l.routeRequests++
				case network_layer.RREP:
					l.routeReplies++
				case network_layer.SESS:
					l.routeSessionData++
				case network_layer.RERR:
					l.routeErrors++
				}
			}
		}

	case simulator.SEND_MSG:
		// receiveEvent := e.(*simulator.ReceiveEvent)
		l.packagesSent++
	case simulator.TERMINATE:

		realtime := time.Since(l.realStartTime)
		terminateEvent := e.(*simulator.TerminateEvent)

		fmt.Printf("\n\n----==== STATISTICS ====----\n\n")

		fmt.Printf("Simulated time: \t\t%s\n", terminateEvent.Time())
		fmt.Printf("Real time: \t\t\t%s\n", realtime.Round(10*time.Millisecond))
		fmt.Printf("\n")

		fmt.Printf("Packages Sent: \t\t\t%d\n", l.packagesSent)
		fmt.Printf("Packages Received: \t\t%d\n", l.packagesReceived)
		fmt.Printf("Dropped Packages: \t\t%d\n", l.packagesSent-l.packagesReceived) // Caveat: This also counts messages that were sent, but did not get received before sim was stopped
		fmt.Printf("\n")

		kbReceived := float64(l.bytesReceived) / 1000
		throughput := (kbReceived) / terminateEvent.Time().Seconds()
		kbMetadata := l.metadataBytes / 1000
		fmt.Printf("Data Sent: \t\t\t%.2f kb\n", kbReceived)
		fmt.Printf("Throughput: \t\t\t%.2f kb/s\n", throughput)
		fmt.Printf("Metadata Bytes Sent: \t\t%d kb\n", kbMetadata)
		fmt.Printf("Metadata Bytes Rate: \t\t%.2f%%\n", (float64(l.metadataBytes)/float64(l.bytesReceived))*100)
		fmt.Printf("Metadata Packages Received: \t%d\n", l.metaDataPackets)
		fmt.Printf("Metadata Package Rate: \t\t%.2f%%\n", (float64(l.metaDataPackets)/float64(l.networkLevelPackages))*100)
		fmt.Printf("\n")

		fmt.Printf("RREQ Packets \t\t\t%d\t(%.2f%%)\n", l.routeRequests, (float64(l.routeRequests)/float64(l.networkLevelPackages))*100)
		fmt.Printf("RREP Packets \t\t\t%d\t(%.2f%%)\n", l.routeReplies, (float64(l.routeReplies)/float64(l.networkLevelPackages))*100)
		fmt.Printf("SESS Packets \t\t\t%d\t(%.2f%%)\n", l.routeSessionData, (float64(l.routeSessionData)/float64(l.networkLevelPackages))*100)
		fmt.Printf("RERR Packets \t\t\t%d\t(%.2f%%)\n", l.routeErrors, (float64(l.routeErrors)/float64(l.networkLevelPackages))*100)
	default:
		return
	}
}

func (l *StatisticsLogger) Log(str string) {}
