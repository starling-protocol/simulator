package starling_node

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/starling"
	"github.com/starling-protocol/starling/device"
)

type Node struct {
	proto     *starling.Protocol
	peers     map[device.DeviceAddress]simulator.Peer
	scenarios []Scenario

	id                   simulator.NodeID
	sim                  *simulator.NodeArguments
	transmissionBehavior simulator.TransmissionBehavior
	random               *rand.Rand

	beforeStartLogs []string
}

func NewNode(random *rand.Rand, id *simulator.NodeID, transmissionBehavior simulator.TransmissionBehavior, options *device.ProtocolOptions) *Node {
	if id == nil {
		randomID := simulator.NodeID(random.Int63())
		id = &randomID
	}

	node := &Node{
		proto:                nil,
		peers:                make(map[device.DeviceAddress]simulator.Peer),
		id:                   *id,
		transmissionBehavior: transmissionBehavior,
		random:               random,
		beforeStartLogs:      []string{},
	}

	node.proto = starling.NewProtocol(NewNodeDevice(node), options)

	return node
}

func idToAddr(id simulator.NodeID) device.DeviceAddress {
	return device.DeviceAddress(fmt.Sprint(id))
}

func (n *Node) UpdateTransmissionBehaviour(transmissionBehavior simulator.TransmissionBehavior) {
	n.transmissionBehavior = transmissionBehavior
}

func (n *Node) OnConnect(peer simulator.Peer, id simulator.NodeID) {
	n.peers[idToAddr(id)] = peer
	n.proto.OnConnection(idToAddr(id))

	for _, scenario := range n.scenarios {
		scenario.OnConnect(n, idToAddr(id))
	}
}

func (n *Node) OnDisconnect(peer simulator.Peer, id simulator.NodeID) {
	delete(n.peers, idToAddr(id))
	n.proto.OnDisconnection(idToAddr(id))

	for _, scenario := range n.scenarios {
		scenario.OnDisconnect(n, idToAddr(id))
	}
}

func (n *Node) OnReceivePacket(peer simulator.Peer, packet []byte, id simulator.NodeID) {
	n.proto.ReceivePacket(idToAddr(id), packet)

	for _, scenario := range n.scenarios {
		scenario.OnReceivePacket(n, packet, idToAddr(id))
	}
}

func (n *Node) OnReceiveData(session device.SessionID, message []byte) {
	for _, scenario := range n.scenarios {
		scenario.OnReceiveData(n, message, session)
	}
}

func (n *Node) OnStart(sim simulator.NodeArguments) {
	n.sim = &sim

	for _, message := range n.beforeStartLogs {
		n.log(message)
	}

	for _, scenario := range n.scenarios {
		scenario.OnStart(n)
	}
}

func (n *Node) ID() simulator.NodeID {
	return n.id
}

func (n *Node) UpdateID(id simulator.NodeID) {
	n.id = id
	n.sim.UpdateID(id)
}

func (n *Node) OnTerminate() {
	for _, scenario := range n.scenarios {
		scenario.OnTerminate(n)
	}
}

func (n *Node) TransmissionBehavior() simulator.TransmissionBehavior {
	return n.transmissionBehavior
}

func (n *Node) SimulatorTime() time.Time {
	return n.sim.Now()
}

func (n *Node) logf(format string, args ...interface{}) {
	n.log(fmt.Sprintf(format, args...))
}

func (n *Node) log(message string) {
	if n.sim == nil {
		n.beforeStartLogs = append(n.beforeStartLogs, message)
		return
	}

	for _, scenario := range n.scenarios {
		scenario.OnLog(n, message)
	}

	n.sim.Log(message)
}

func (n *Node) AddScenario(scenario Scenario) {
	n.scenarios = append(n.scenarios, scenario)
}

func makeSharedSecret(random *rand.Rand) device.SharedSecret {
	secret := [32]byte{}

	random.Read(secret[:])
	return secret[:]
}

func LinkNodes(a *Node, b *Node) device.ContactID {
	sessA, err := a.proto.LinkingStart()
	if err != nil {
		panic(err)
	}

	sessB, err := b.proto.LinkingStart()
	if err != nil {
		panic(err)
	}

	contactA, err := a.proto.LinkingCreate(sessA, sessB.GetShare())
	if err != nil {
		panic(err)
	}

	contactB, err := b.proto.LinkingCreate(sessB, sessA.GetShare())
	if err != nil {
		panic(err)
	}

	if contactA != contactB {
		panic("contact IDs are different")
	}

	return contactA
}

func GroupLinkNodes(nodes []*Node) device.ContactID {
	if len(nodes) < 1 {
		panic("Cannot create a group with 0 members")
	}
	var contact device.ContactID
	secret := makeSharedSecret(nodes[0].random)
	for _, node := range nodes {
		var err error
		contact, err = node.proto.JoinGroup(secret)
		if err != nil {
			panic(err)
		}
	}
	return contact
}
