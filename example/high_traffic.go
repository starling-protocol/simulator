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
	"github.com/starling-protocol/simulator/visualizer"
	"github.com/starling-protocol/starling/utils"
)

func highTrafficExample() {
	seed := 11 // rand.Int63()
	fmt.Printf("Seed: %d\n", seed)
	random := rand.New(rand.NewSource(int64(seed)))

	var transmissionBehavior = transmission_behavior.RandomDrops{
		DropChance: 0.01,
		Delay:      time.Duration(20) * time.Millisecond,
		Random:     random,
	}

	var loggerList []simulator.Logger
	// mainLogger := loggers.NewStandardLogger()
	// loggerList = append(loggerList, mainLogger)
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

	for i := 0; i < 10; i++ {
		a := random.Intn(len(nodes))
		b := random.Intn(len(nodes))

		contact := node.LinkNodes(nodes[a], nodes[b])
		orchestrator := node.NewSessionOrchestrator(contact)

		sessionA := orchestrator.SessionScenario(true)
		nodes[a].AddScenario(sessionA)

		sessionB := orchestrator.SessionScenario(true)
		nodes[b].AddScenario(sessionB)

		lorem1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus ullamcorper euismod mauris. Phasellus aliquet consequat ex. Aenean id luctus eros. Nulla eu dolor suscipit odio porta elementum eget eu elit. Etiam tincidunt fringilla elit in porttitor. Fusce sit amet quam ultricies, porta enim eu, ultricies enim. Nunc id nisi id eros tincidunt tempor a nec est. Vestibulum nec rhoncus justo, ac placerat turpis. Pellentesque lobortis fringilla nisl non feugiat. Pellentesque congue ultrices risus, consequat vestibulum nulla condimentum eu. Vivamus vel libero sollicitudin tortor bibendum euismod ac sed ligula. Fusce nec risus felis. Quisque leo massa, finibus sed condimentum vel, pretium sed orci."
		lorem2 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc placerat aliquam lobortis. Aenean ornare bibendum lacus, eu lacinia enim lobortis vitae. Donec placerat lobortis turpis, vel imperdiet sapien luctus non. Aliquam augue neque, congue eu lobortis nec, pretium ac justo. Nam justo sapien, porta eget sodales quis, auctor non mauris. In sagittis condimentum diam, sed fringilla turpis vestibulum quis. Etiam nisi tellus, posuere ut velit vel, tincidunt volutpat justo. Maecenas convallis sem magna, venenatis fermentum augue iaculis eget. Aliquam non justo sit amet ipsum laoreet ornare. Nulla vitae eros mollis neque tincidunt placerat sed quis est. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus."

		r := random.Intn(200000) + 10000
		delay := time.Duration(r) * time.Millisecond

		for i := 0; i < 100; i++ {
			if i == 0 {
				nodes[a].AddScenario(node.DelayScenario(sessionA.SendDataScenario(fmt.Sprintf("%d", i)+lorem1), delay))
			} else {
				nodes[a].AddScenario(node.EventScenario(sessionA.ReceiveDataEvent(fmt.Sprintf("%d", i-1) + lorem2)).
					OnEvent(node.DelayScenario(sessionA.SendDataScenario(fmt.Sprintf("%d", i)+lorem1), delay)))
			}
			nodes[b].AddScenario(
				node.EventScenario(sessionB.ReceiveDataEvent(fmt.Sprintf("%d", i) + lorem1)).
					OnEvent(sessionB.SendDataScenario(fmt.Sprintf("%d", i) + lorem2)),
			)
		}
		nodes[a].AddScenario(node.EventScenario(sessionA.ReceiveDataEvent(fmt.Sprintf("%d", 99) + lorem2)).
			OnEvent(node.ActionScenario(func(node *node.Node) {
				count++
				// fmt.Printf("%d: Received pong\n", count)
			})))
	}

	visualizer.StartGUI(sim, false, 0.05, "../../evaluation/scenarios/vadere_data/protest.scenario")

	// sim.Update(10 * time.Second)
	// sim.Terminate()

}
