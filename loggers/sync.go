package loggers

import (
	"fmt"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/starling_node"
)

type SyncLogger struct {
	realStartTime      time.Time
	syncStateScenarios []*starling_node.ScenarioStoreSyncState
}

func NewSyncLogger() *SyncLogger {
	return &SyncLogger{}
}

func (l *SyncLogger) SetSyncStateScenario(syncStateScenarios []*starling_node.ScenarioStoreSyncState) {
	l.syncStateScenarios = syncStateScenarios
}

func (l *SyncLogger) Init() {
	l.realStartTime = time.Now()
}

func (l *SyncLogger) NewEvent(e simulator.Event) {

	switch e.EventType() {
	case simulator.TERMINATE:

		nodeLines := []string{}
		totalMessageSentGoal := 0
		totalReceived := 0
		for i, scenario := range l.syncStateScenarios {
			state := scenario.GetSyncState()
			syncGoal := scenario.GetSyncGoal()
			totalMessageSentGoal += syncGoal.MessagesSent

			messagesReceived := 0
			ratio := 1.0
			if state != nil {
				for _, nodeState := range state.NodeStates {
					messagesReceived += len(nodeState)
				}
				totalReceived += messagesReceived
				if messagesReceived > syncGoal.MessagesSent {
					panic("the amount of received messages cannot be larger than the message goal")
				}
				if syncGoal.MessagesSent > 0 {
					ratio = float64(messagesReceived) / float64(syncGoal.MessagesSent)
				}
			} else {
				ratio = 0
			}
			nodeLines = append(nodeLines, fmt.Sprintf("node: %d, received %d msgs out of %d \t(%.2f%%)\n", i, messagesReceived, syncGoal.MessagesSent, ratio*100))
		}

		realtime := time.Since(l.realStartTime)
		terminateEvent := e.(*simulator.TerminateEvent)

		fmt.Printf("\n\n----==== STATISTICS ====----\n\n")

		// fmt.Printf("Packages Sent: \t\t\t%d\n", l.packagesSent)
		for _, line := range nodeLines {
			fmt.Print(line)
		}

		fmt.Printf("Simulated time: \t%s\n", terminateEvent.Time())
		fmt.Printf("Real time: \t\t%s\n", realtime.Round(10*time.Millisecond))
		fmt.Printf("\n")

		fmt.Printf("Total messages received: \t%d\n", totalReceived)
		fmt.Printf("Total message received goal: \t%d\n", totalMessageSentGoal)
		ratio := 1.0
		if totalMessageSentGoal > 0 {
			ratio = float64(totalReceived) / float64(totalMessageSentGoal)
		}
		fmt.Printf("Received to goal ratio: \t%2f%%\n", ratio*100)

	default:
		return
	}
}

func (l *SyncLogger) Log(str string) {}
