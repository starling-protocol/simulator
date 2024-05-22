package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/loggers"
	node "github.com/starling-protocol/simulator/starling_node"
	"github.com/starling-protocol/simulator/transmission_behavior"
	"github.com/starling-protocol/simulator/vadere"
	"github.com/starling-protocol/starling/utils"
)

func vaderExample() {
	seed := 11 // rand.Int63()
	fmt.Printf("Seed: %d\n", seed)
	random := rand.New(rand.NewSource(int64(seed)))

	var transmissionBehavior = transmission_behavior.RandomDrops{
		DropChance: 0,
		Delay:      time.Duration(20) * time.Millisecond,
		Random:     random,
	}

	var loggerList []simulator.Logger
	mainLogger := loggers.NewStandardLogger()
	loggerList = append(loggerList, mainLogger)
	// pcapLogger := &loggers.PCAPLogger{}
	// loggerList = append(loggerList, pcapLogger)
	statisticsLogger := &loggers.StatisticsLogger{}
	loggerList = append(loggerList, statisticsLogger)
	profileLogger := &loggers.ProfileLogger{}
	loggerList = append(loggerList, profileLogger)

	var sim = simulator.NewSimulator(20.0, time.Duration(20*time.Millisecond), random, loggerList)

	vadereNodeMap := vadere.VadereLoad("../vadere/old/postvis.traj", 0.1, random)
	nodes := []*node.Node{}

	for i, nodeId := range utils.ShuffleMapKeys(random, vadereNodeMap) {
		vadereNode, found := vadereNodeMap[nodeId]
		if !found {
			panic("vadere node not found")
		}
		var vadereProfile = vadere.NewMovementProfile(vadereNode.CoordList)
		var n *node.Node = nil

		if i == 0 {
			n = node.NewNode(random, (*simulator.NodeID)(&nodeId), transmissionBehavior, nil)
		} else if i == 1 {
			n = node.NewNode(random, (*simulator.NodeID)(&nodeId), transmissionBehavior, nil)
		} else {
			n = node.NewNode(random, (*simulator.NodeID)(&nodeId), transmissionBehavior, nil)
		}

		sim.AddNode(n, vadereProfile, *vadereNode.StartTime)
		nodes = append(nodes, n)
		node.NetworkLayerColors(n)
	}

	count := 0

	for i := 0; i < 100; i++ {
		a := random.Intn(len(nodes))
		b := random.Intn(len(nodes))
		if a == b {
			continue
		}

		contact := node.LinkNodes(nodes[a], nodes[b])
		orchestrator := node.NewSessionOrchestrator(contact)

		sendDelay := time.Duration(random.Intn(int(200*time.Second))) + time.Second
		sessionA := orchestrator.SessionScenario(true)
		nodes[a].AddScenario(sessionA)
		sessionB := orchestrator.SessionScenario(true)
		nodes[b].AddScenario(sessionB)

		nodes[a].AddScenario(node.DelayScenario(sessionA.SendDataScenario("ping"), sendDelay))

		nodes[b].AddScenario(
			node.EventScenario(sessionB.ReceiveDataEvent("ping")).
				OnEvent(sessionB.SendDataScenario("pong")),
		)

		nodes[a].AddScenario(node.EventScenario(sessionA.ReceiveDataEvent("pong")).
			OnEvent(node.ActionScenario(func(node *node.Node) {
				count++
				// fmt.Printf("%d: Received pong\n", count)
			})))

	}

	// visualizer.StartGUI(sim, false, 0.05, true)

	sim.Update(20 * time.Second)
	sim.Terminate()

}
