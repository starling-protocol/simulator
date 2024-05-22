package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/loggers"
	"github.com/starling-protocol/simulator/movement_profiles"
	node "github.com/starling-protocol/simulator/starling_node"
	"github.com/starling-protocol/simulator/transmission_behavior"
	"github.com/starling-protocol/simulator/visualizer"

	"github.com/starling-protocol/starling/device"
)

func syncExample() {
	seed := 20 //rand.Int63()
	fmt.Printf("Seed: %d\n", seed)
	random := rand.New(rand.NewSource(int64(seed)))

	var transmissionBehavior = transmission_behavior.RandomDrops{
		DropChance: 0.01,
		Delay:      time.Duration(20) * time.Millisecond,
		Random:     random,
	}

	// var transmissionBehavior = transmissionbehavior.DelayTransmission{
	// 	Delay: 20 * time.Millisecond,
	// }

	var movementProfile = movement_profiles.NewRandomNode(50, 50, 60, 20, random)

	var loggerList []simulator.Logger
	logger := loggers.NewStandardLogger()
	loggerList = append(loggerList, logger)
	syncLogger := &loggers.SyncLogger{}
	loggerList = append(loggerList, syncLogger)

	var sim = simulator.NewSimulator(20.0, time.Duration(20*time.Millisecond), random, loggerList)

	protocolOptions := device.DefaultSyncProtocolOptions()

	nodes := []*node.Node{}

	// Create nodes
	for i := 0; i < 20; i++ {
		var n *node.Node = nil

		nodeId := (simulator.NodeID)(int64(i))
		n = node.NewNode(random, &nodeId, transmissionBehavior, protocolOptions)

		sim.AddNode(n, movementProfile, 0)
		node.NetworkLayerColors(n)
		nodes = append(nodes, n)
	}

	syncStateScenarios := []*node.ScenarioStoreSyncState{}

	// Create groups
	for i := 0; i < 10; i++ {
		groupSize := random.Intn(10) // TODO: Fix this
		group := []*node.Node{}
		for i := 0; i < groupSize; i++ {
			randomNode := nodes[random.Intn(len(nodes))]
			if !slices.Contains(group, randomNode) {
				group = append(group, randomNode)
			}
		}

		contact := node.GroupLinkNodes(group)

		// Create messages in the group
		totalGroupMessageCount := 0
		for _, n := range group {
			node.SyncModelDisplay(n, contact)

			messageCount := random.Intn(10)
			totalGroupMessageCount += messageCount

			for i := 0; i < messageCount; i++ {
				r := random.Intn(60) + 1
				delay := time.Duration(r) * time.Second
				// n.AddScenario(node.AddMessageAndSynchronizeScenario(contact, []byte(fmt.Sprintf("Message %d from %d", i, n.ID())), nil, &delay))
				n.AddScenario(node.DelayScenario(node.SyncAddMessageScenario(contact, []byte(fmt.Sprintf("Message %d from %d", i, n.ID())), nil), delay))
				// n.AddScenario(node.DelayScenario(node.SyncContactScenario(contact), (time.Duration(r)+1)*time.Second))
			}
		}

		syncGoal := node.NewSyncGoal(i, totalGroupMessageCount, totalGroupMessageCount*groupSize, groupSize)

		for _, n := range group {
			syncStateScenario := node.StoreSyncStateScenario(contact, syncGoal, int(n.ID()))
			syncStateScenarios = append(syncStateScenarios, syncStateScenario)
			n.AddScenario(syncStateScenario)
		}
	}

	syncLogger.SetSyncStateScenario(syncStateScenarios)

	visualizer.StartGUI(sim, false, 1, "")
	// sim.Update(120 * time.Second)
	// sim.Terminate()
}
