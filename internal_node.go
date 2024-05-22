package simulator

import (
	"math/rand"
	"time"

	"github.com/starling-protocol/starling/utils"
)

type InternalNode struct {
	node                Node
	nodeID              NodeID
	internalID          InternalID
	coords              Coordinate
	movementInstruction MovementInstruction
	nodeMovement        NodeMovement
	peers               map[InternalID]Peer
	random              *rand.Rand
	sim                 *Simulator
	curBufferCount      int
	bufferSize          int
	lastMessageSent     time.Duration
	data                map[string]interface{}
}

type InternalID int

type NodeArguments struct {
	Data      func() *map[string]interface{}
	UpdateID  func(newId NodeID)
	DelayBy   func(func(), time.Duration)
	Now       func() time.Time
	Terminate func(error)
	Log       func(message string)
	Loggers   []Logger
}

func (n *InternalNode) Node() Node {
	return n.node
}

func (n *InternalNode) Coords() Coordinate {
	return n.coords
}

func (n *InternalNode) Data() *map[string]interface{} {
	return &n.data
}

func (n *InternalNode) NodeID() NodeID {
	return n.nodeID
}

func (nodeA *InternalNode) connectNodes(nodeB *InternalNode) {
	peer := Peer{
		target: nodeB,
		origin: nodeA,
	}
	nodeA.peers[nodeB.internalID] = peer
	nodeA.node.OnConnect(peer, nodeB.nodeID)

}

func (nodeA *InternalNode) disconnectNodes(nodeB *InternalNode) {
	peer, found := nodeA.getPeer(nodeB)
	if found {
		delete(nodeA.peers, nodeB.internalID)
		nodeA.node.OnDisconnect(*peer, nodeB.nodeID)
	} else {
		panic("could not find node to disconnect in internalNode.disconnectNodes")
	}
}

func (nodeA *InternalNode) getPeer(nodeB *InternalNode) (*Peer, bool) {
	peer, found := nodeA.peers[nodeB.internalID]
	if found {
		return &peer, true
	} else {
		return nil, false
	}
}

func getPeerIndex(peers []InternalPeer, nodeA *InternalNode, nodeB *InternalNode) int {
	for i, peer := range peers {
		if peer.a == nodeA {
			if peer.b == nodeB {
				return i
			}
		} else if peer.a == nodeB {
			if peer.b == nodeA {
				return i
			}
		}
	}
	return -1
}

func removePeer(peers []InternalPeer, nodeA *InternalNode, nodeB *InternalNode) []InternalPeer {
	index := getPeerIndex(peers, nodeA, nodeB)
	if index != -1 {
		peers[index] = peers[len(peers)-1]
		return peers[:len(peers)-1]
	} else {
		return peers
	}
}

func (n *InternalNode) updateData(key string, data interface{}) {
	n.data[key] = data
}

func (n *InternalNode) updateID(nodeID NodeID) {
	n.nodeID = nodeID

	for _, internalID := range utils.ShuffleMapKeys(n.random, n.peers) {
		n.peers[internalID].target.disconnectNodes(n)
		n.disconnectNodes(n.peers[internalID].target)
	}
	// TODO: Update all nodes that has it as peer, such that the old one disconnects and the new one connects
}

func (n *InternalNode) delayBy(functionToCall func(), delay time.Duration) {
	n.sim.pushDelayEvent(n.sim.time+delay, n, functionToCall)
}

func (n *InternalNode) terminate(err error) {
	n.sim.pushTerminateEvent(n.sim.time, err)
}

func (n *InternalNode) log(message string) {
	for _, logger := range n.sim.loggers {
		logger.Log(message)
	}
}

func (n *InternalNode) startNode() {
	nodeArgs := NodeArguments{
		Data:      n.Data,
		UpdateID:  n.updateID,
		DelayBy:   n.delayBy,
		Now:       n.sim.Now,
		Terminate: n.terminate,
		Log:       n.log,
	}
	n.node.OnStart(nodeArgs)
}

func (n *InternalNode) hasNeighbour(target *InternalNode) bool {
	_, found := n.getPeer(target)
	return found
}
