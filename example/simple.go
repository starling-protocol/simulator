package main

import (
	"fmt"
	"math/rand"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/loggers"
	"github.com/starling-protocol/simulator/movement_profiles"
	node "github.com/starling-protocol/simulator/starling_node"
	"github.com/starling-protocol/simulator/transmission_behavior"
	"github.com/starling-protocol/simulator/visualizer"
)

func simpleExample() {
	seed := 21 //rand.Int63()
	fmt.Printf("Seed: %d\n", seed)
	random := rand.New(rand.NewSource(int64(seed)))

	// var transmissionBehavior = transmission_behavior.BleTransmission{}

	// var transmissionBehavior = transmission_behavior.DelayTransmission{
	// 	Delay: 20 * time.Millisecond,
	// }
	var transmissionBehavior = transmission_behavior.LongTailTransmission{
		Alpha:      2.7,
		X_m:        20,
		DropChance: 0.0,
		Random:     *random,
	}

	var movementProfile = movement_profiles.NewRandomNode(50, 50, 60, 20, random)

	var loggerList []simulator.Logger
	logger := loggers.NewStandardLogger()
	loggerList = append(loggerList, logger)
	pcapLogger := &loggers.PCAPLogger{}
	loggerList = append(loggerList, pcapLogger)

	var sim = simulator.NewSimulator(20.0, 20, random, loggerList)

	nodes := []*node.Node{}

	for i := 0; i < 50; i++ {
		var n *node.Node = nil

		nodeId := (simulator.NodeID)(int64(i))
		n = node.NewNode(random, &nodeId, transmissionBehavior, nil)
		node.NetworkLayerColors(n)

		sim.AddNode(n, movementProfile, 0)
		nodes = append(nodes, n)
	}

	contact := node.LinkNodes(nodes[0], nodes[1])

	orchestrator := node.NewSessionOrchestrator(contact)
	sessionA := orchestrator.SessionScenario(true)
	nodes[0].AddScenario(sessionA)
	sessionB := orchestrator.SessionScenario(true)
	nodes[1].AddScenario(sessionB)

	nodes[0].AddScenario(sessionA.SendDataScenario("ping"))

	nodes[1].AddScenario(
		node.EventScenario(sessionB.ReceiveDataEvent("ping")).
			OnEvent(sessionB.SendDataScenario("pong")),
	)

	// nodes[0].AddScenario(
	// 	node.EventScenario(sessionA.ReceiveDataEvent("pong")).
	// 		OnEvent(node.TerminateScenario()),
	// )

	// sim.Update(30 * time.Second)
	// sim.Terminate()

	visualizer.StartGUI(sim, false, 1, "")
}
