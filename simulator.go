package simulator

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Coordinate struct {
	X float64
	Y float64
}

func (a *Coordinate) Distance(b Coordinate) float64 {
	dX := a.X - b.X
	dY := a.Y - b.Y
	return math.Sqrt(dX*dX + dY*dY)
}

type MovementInstruction struct {
	Coords Coordinate
	Time   time.Duration
}

func NewMovementInstruction(x float64, y float64, time time.Duration) MovementInstruction {
	return MovementInstruction{
		Coords: Coordinate{X: x, Y: y},
		Time:   time,
	}
}

type NodeMovement interface {
	StartPosition() Coordinate
	RegisterMovements(Coordinate) MovementInstruction
}

type TransmissionBehavior interface {
	Transmission(Coordinate, Coordinate, []byte) (shouldBeDropped bool, delay time.Duration)
}

type NodeID int64

type Node interface {
	OnConnect(peer Peer, id NodeID)
	OnDisconnect(peer Peer, id NodeID)
	OnReceivePacket(peer Peer, packet []byte, id NodeID)
	OnStart(sim NodeArguments)
	ID() NodeID
	OnTerminate()
	TransmissionBehavior() TransmissionBehavior
}

type Peer struct {
	target *InternalNode
	origin *InternalNode
}

func (p Peer) SendPacket(packet []byte) {
	sim := p.origin.sim

	if p.origin.curBufferCount+1 > p.origin.bufferSize {
		sim.logDebug(fmt.Sprintf("simulator:packet:buffer_full:%d:%d", p.origin.nodeID, p.target.nodeID))
		return
	}

	transmissionBehavior := p.origin.node.TransmissionBehavior()
	shouldBeDropped, propagationDelay := transmissionBehavior.Transmission(p.origin.coords, p.target.coords, packet)

	p.origin.curBufferCount++

	var sendTime time.Duration
	if p.origin.lastMessageSent < sim.time {
		sendTime = sim.time + sim.transmissionDelay
	} else {
		sendTime = p.origin.lastMessageSent + sim.transmissionDelay
	}

	sim.pushSendEvent(sendTime, p.target, p.origin, p.target.nodeID, p, shouldBeDropped, propagationDelay, packet)
	p.origin.lastMessageSent = sendTime + propagationDelay // This represents the time at which the next message could begin sending
}

func (p *Peer) Data() *map[string]any {
	if p.origin.hasNeighbour(p.target) {
		return p.origin.sim.getPeerData(p, p.origin)
	}
	return nil
}

type Logger interface {
	Log(message string)
	NewEvent(event Event)
	Init()
}

type InternalPeer struct {
	a     *InternalNode
	b     *InternalNode
	dataA map[string]interface{}
	dataB map[string]interface{}
}

func (p *InternalPeer) NodeA() *InternalNode {
	return p.a
}

func (p *InternalPeer) NodeB() *InternalNode {
	return p.b
}

func (p *InternalPeer) Data(node *InternalNode) *map[string]any {
	if node == p.a {
		return &p.dataA
	} else if node == p.b {

		return &p.dataB
	} else {
		return nil
	}
}

type Simulator struct {
	eventQueue            *EventQueue
	nodes                 []*InternalNode
	peers                 []InternalPeer
	isRunning             bool
	nodeRange             float64 // nodeRange is the actual node range squared (for more efficient distance computation)
	transmissionDelay     time.Duration
	time                  time.Duration
	random                *rand.Rand
	currentSequenceNumber int64
	loggers               []Logger
	lastEvent             Event
	regionMap             *RegionMap
	isTerminating         bool
}

func NewSimulator(bleRange float64, transmissionDelay time.Duration, random *rand.Rand, loggers []Logger) *Simulator {
	pq := make(EventQueue, 0)
	heap.Init(&pq)

	return &Simulator{
		eventQueue:            &pq,
		nodes:                 []*InternalNode{},
		peers:                 []InternalPeer{},
		isRunning:             false,
		nodeRange:             bleRange * bleRange,
		transmissionDelay:     transmissionDelay,
		time:                  0,
		random:                random,
		currentSequenceNumber: 0,
		loggers:               loggers,
		lastEvent:             nil,
		regionMap:             NewRegionMap(bleRange),
		isTerminating:         false,
	}
}

// Start will start the simulator if it is not already running
func (s *Simulator) Start() {
	if !s.isRunning {
		s.isRunning = true

		// Push initial event
		s.pushTimeStepEvent(0)
		s.lastEvent = (*s.eventQueue)[0]
		s.updateLoggers(s.lastEvent)

		for _, logger := range s.loggers {
			logger.Init()
		}

		for _, iNode := range s.nodes {
			iNode.startNode()
			iNode.nodeID = iNode.node.ID()
		}
	}
}

func (s *Simulator) IsRunning() bool {
	return s.isRunning
}

func (s *Simulator) IsTerminating() bool {
	return s.isTerminating
}

func (s *Simulator) Nodes() []*InternalNode {
	return s.nodes
}

func (s *Simulator) Peers() []InternalPeer {
	return s.peers
}

func (s *Simulator) Now() time.Time {
	return time.Unix(0, 0).Add(s.time)
}

func (s *Simulator) Update(updateUntilTime time.Duration) error {
	if !s.isRunning {
		s.Start()
	}

	if s.eventQueue.Len() > 0 {

		e := heap.Pop(s.eventQueue).(Event)

		for {
			if e.Time() > updateUntilTime {
				heap.Push(s.eventQueue, e)
				break
			}

			s.time = e.Time()
			s.updateLoggers(e)
			s.lastEvent = e

			switch e.EventType() {
			case TIMESTEP:
				s.updateLocations()
				s.pushTimeStepEvent(e.Time() + 10*time.Millisecond)
			case CONNECT:
				connectEvent := e.(*ConnectEvent)
				s.connectNodes(connectEvent.nodeA, connectEvent.nodeB)
			case DISCONNECT:
				disconnectEvent := e.(*DisconnectEvent)
				s.disconnectNodes(disconnectEvent.nodeA, disconnectEvent.nodeB)
			case ADD_NODE:
				addNodeEvent := e.(*AddNodeEvent)
				iNode := addNodeEvent.node
				s.nodes = append(s.nodes, iNode)
				s.regionMap.AddNode(iNode)
				iNode.startNode()
				iNode.nodeID = iNode.node.ID()
			case RCV_MSG:
				receiveEvent := e.(*ReceiveEvent)
				receiveEvent.origin.curBufferCount--
				if receiveEvent.target.hasNeighbour(receiveEvent.origin) {
					packet := receiveEvent.packet
					receiveEvent.target.node.OnReceivePacket(receiveEvent.peer, packet, receiveEvent.originNodeID)
				}
			case SEND_MSG:
				sendEvent := e.(*SendEvent)
				s.sendPacket(sendEvent)
			case DELAY:
				delayEvent := e.(*DelayEvent)
				delayEvent.functionToCall()
			case TERMINATE:
				s.isTerminating = true
				terminateEvent := e.(*TerminateEvent)
				if terminateEvent.err == nil {
					for _, n := range s.nodes {
						n.node.OnTerminate()
					}
				}
				s.isRunning = false
				return terminateEvent.err
			default:
				panic("Simulator event error!")
			}

			if s.eventQueue.Len() == 0 {
				break
			} else {
				e = heap.Pop(s.eventQueue).(Event)
			}
		}
	}
	return nil
}

func (s *Simulator) updateLocations() {
	deltaTime := 10 * time.Millisecond

	// Update locations
	for i := 0; i < len(s.nodes); i++ {
		if s.nodes[i].movementInstruction.Time <= 0 {
			s.nodes[i].movementInstruction = s.nodes[i].nodeMovement.RegisterMovements(s.nodes[i].coords)
		} else {
			node := s.nodes[i]
			var delta_x, delta_y = calcSpeed(node, node.movementInstruction.Coords, node.movementInstruction.Time, time.Duration(deltaTime))
			newCoords := Coordinate{node.coords.X + delta_x, node.coords.Y + delta_y}
			s.regionMap.MoveNode(node, node.coords, newCoords)
			node.coords = newCoords

			s.nodes[i].movementInstruction.Time -= time.Duration(deltaTime)
		}
	}

	disconnects := []*DisconnectEvent{}
	connects := []*ConnectEvent{}

	// Update connections
	for i := 0; i < len(s.nodes); i++ {
		nodeA := s.nodes[i]
		relevantNodes, oldPeers := s.regionMap.WithinRange(nodeA)

		for _, oldPeer := range oldPeers {

			shouldDisconnect := true
			for _, e := range disconnects {
				// Check if nodeA has already been part of a disconnect event
				if e.nodeB.internalID == nodeA.internalID && e.nodeA.internalID == oldPeer.internalID {
					shouldDisconnect = false
				}
			}
			if shouldDisconnect {
				// Remove peer
				s.peers = removePeer(s.peers, nodeA, oldPeer)

				// Add disconnect event to queue
				s.pushDisconnectEvent(s.time+deltaTime, nodeA, oldPeer)

				// Store disconnect event, so we don't add a symmetrical event later
				disconnects = append(disconnects, s.NewDisconnectEvent(s.time+deltaTime, nodeA, oldPeer))
			}
		}

		for _, nodeB := range relevantNodes {
			hasPeer := nodeA.hasNeighbour(nodeB)
			if !hasPeer {

				shouldConnect := true
				for _, e := range connects {
					// Check if nodeA has already been part of a disconnect event
					if e.nodeB.internalID == nodeA.internalID && e.nodeA.internalID == nodeB.internalID {
						shouldConnect = false
					}
				}

				if shouldConnect {
					// Add peer
					simPeer := InternalPeer{nodeA, nodeB, make(map[string]interface{}), make(map[string]interface{})}
					s.peers = append(s.peers, simPeer)

					// Add connect event to queue
					s.pushConnectEvent(s.time+deltaTime, nodeA, nodeB)

					connects = append(connects, s.NewConnectEvent(s.time+deltaTime, nodeA, nodeB))
				}
			}
		}
	}
}

func (s *Simulator) connectNodes(nodeA *InternalNode, nodeB *InternalNode) {
	nodeA.connectNodes(nodeB)
	nodeB.connectNodes(nodeA)
}

func (s *Simulator) disconnectNodes(nodeA *InternalNode, nodeB *InternalNode) {
	nodeA.disconnectNodes(nodeB)
	nodeB.disconnectNodes(nodeA)
}

func (s *Simulator) sendPacket(sendEvent *SendEvent) {
	peer := sendEvent.peer

	if !sendEvent.shouldBeDropped && peer.origin.hasNeighbour(peer.target) {
		s.logDebug(fmt.Sprintf("simulator:packet:send:%d:%d", peer.origin.nodeID, peer.target.nodeID))

		newPeer := Peer{
			target: peer.origin,
			origin: peer.target,
		}

		s.pushReceiveEvent(s.time+sendEvent.delay, peer.target, peer.origin, peer.origin.nodeID, newPeer, sendEvent.packet)
	} else {
		peer.origin.curBufferCount--
		s.logDebug(fmt.Sprintf("simulator:packet:drop:%d:%d", peer.origin.nodeID, peer.target.nodeID))
	}
}

func (s *Simulator) Terminate() {
	s.isTerminating = true
	s.pushTerminateEvent(s.time, nil)
	s.Update(s.time)
}

func (s *Simulator) AddNode(node Node, nodeMovement NodeMovement, delay time.Duration) {
	var iNode = &InternalNode{
		node:                node,
		nodeID:              node.ID(),
		internalID:          InternalID(s.random.Int()),
		coords:              nodeMovement.StartPosition(),
		movementInstruction: MovementInstruction{Coordinate{0, 0}, -1},
		nodeMovement:        nodeMovement,
		peers:               make(map[InternalID]Peer),
		random:              s.random,
		sim:                 s,
		curBufferCount:      0,
		bufferSize:          400,
		lastMessageSent:     time.Duration(0),
		data:                make(map[string]interface{}),
	}

	s.pushAddNodeEvent(delay, iNode)
}

func nodeDistSquared(nodeA *InternalNode, nodeB *InternalNode) float64 {
	diffX := math.Abs(float64(nodeA.coords.X) - float64(nodeB.coords.X))
	diffY := math.Abs(float64(nodeA.coords.Y) - float64(nodeB.coords.Y))

	return diffX*diffX + diffY*diffY
}

func calcSpeed(node *InternalNode, coord Coordinate, time time.Duration, timeSinceLast time.Duration) (float64, float64) {
	if int64(time) == 0 {
		return 0, 0
	}
	var delta_x = ((coord.X - node.coords.X) / (float64(time) / 1000000000.0)) * (float64(timeSinceLast) / 1000000000.0)
	var delta_y = ((coord.Y - node.coords.Y) / (float64(time) / 1000000000.0)) * (float64(timeSinceLast) / 1000000000.0)
	return delta_x, delta_y
}

func (s *Simulator) updateLoggers(e Event) {
	for _, logger := range s.loggers {
		logger.NewEvent(e)
	}
}

func (s *Simulator) logDebug(str string) {
	for _, logger := range s.loggers {
		logger.Log(str)
	}
}

func (s *Simulator) getPeerData(peer *Peer, node *InternalNode) *map[string]any {
	i := getPeerIndex(s.peers, peer.origin, peer.target)
	if i < 0 {
		return nil
	}
	if s.peers[i].a == node {
		return &s.peers[i].dataA
	} else if s.peers[i].b == node {
		return &s.peers[i].dataB
	}
	return nil
}
