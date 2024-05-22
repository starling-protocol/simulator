package starling_node

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/starling-protocol/starling/device"
	"github.com/starling-protocol/starling/sync"

	crypto_rand "crypto/rand"
)

type NodeDevice struct {
	node              *Node
	contactsContainer *device.MemoryContactsContainer
}

func NewNodeDevice(node *Node) *NodeDevice {
	return &NodeDevice{
		node:              node,
		contactsContainer: device.NewMemoryContactsContainer(),
	}
}

// Log implements device.Device.
func (dev *NodeDevice) Log(message string) {
	dev.node.log(message)
}

// SendPacket implements device.Device.
func (dev *NodeDevice) SendPacket(address device.DeviceAddress, packet []byte) {
	peer := dev.node.peers[address]
	peer.SendPacket(packet)
}

// MessageDelivered implements device.Device.
func (dev *NodeDevice) MessageDelivered(messageID device.MessageID) {
	for _, scenario := range dev.node.scenarios {
		scenario.OnMessageDelivered(dev.node, messageID)
	}
}

// MaxPacketSize implements device.Device.
func (dev *NodeDevice) MaxPacketSize(address device.DeviceAddress) (int, error) {
	return 517, nil
}

// ProcessMessage implements device.Device.
func (dev *NodeDevice) ProcessMessage(session device.SessionID, message []byte) {
	dev.node.log(fmt.Sprintf("application:receive:%d '%s'", session, string(message)))
	dev.node.OnReceiveData(session, message)
}

// SyncStateChanged implements device.Device.
func (dev *NodeDevice) SyncStateChanged(contact device.ContactID, stateUpdate []byte) {
	data, err := sync.DecodeModelFromJSON(stateUpdate)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal json in SyncStateChanged: %s", err.Error()))
	}

	for _, scenario := range dev.node.scenarios {
		scenario.OnSyncStateChanged(dev.node, contact, *data)
	}
}

// Rand implements device.Device.
func (dev *NodeDevice) Rand() *rand.Rand {
	if dev.node.random == nil {
		panic("Node device 'random' was nil")
	}

	return dev.node.random
}

// CryptoRand implements device.Device.
func (dev *NodeDevice) CryptoRand() io.Reader {
	// return dev.Rand()
	return crypto_rand.Reader
}

// ReplyPayload implements device.Device.
func (*NodeDevice) ReplyPayload(session device.SessionID, contact device.ContactID) []byte {
	return nil
}

// SessionBroken implements device.Device.
func (dev *NodeDevice) SessionBroken(session device.SessionID) {
	for _, scenario := range dev.node.scenarios {
		scenario.OnSessionBroken(dev.node, session)
	}
}

// SessionEstablished implements device.Device.
func (dev *NodeDevice) SessionEstablished(session device.SessionID, contact device.ContactID, address device.DeviceAddress) {
	for _, scenario := range dev.node.scenarios {
		scenario.OnSessionEstablished(dev.node, session, contact, address)
	}
}

// Delay implements device.Device.
func (n *NodeDevice) Delay(action func(), duration time.Duration) {
	n.node.sim.DelayBy(action, duration)
}

// Now implements device.Device.
func (n *NodeDevice) Now() time.Time {
	return n.node.sim.Now()
}

// ContactsContainer implements device.Device.
func (n *NodeDevice) ContactsContainer() device.ContactsContainer {
	return n.contactsContainer
}
