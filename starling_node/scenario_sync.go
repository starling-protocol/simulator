package starling_node

import (
	"github.com/starling-protocol/starling/device"
	"github.com/starling-protocol/starling/sync"
)

type ScenarioSyncAddMessage struct {
	EmptyScenario
	contact         device.ContactID
	message         []byte
	attachedContact *device.ContactID
}

func SyncAddMessageScenario(contact device.ContactID, message []byte, attachedContact *device.ContactID) *ScenarioSyncAddMessage {
	return &ScenarioSyncAddMessage{
		contact:         contact,
		message:         message,
		attachedContact: attachedContact,
	}
}

func (s *ScenarioSyncAddMessage) OnStart(node *Node) {
	node.proto.SyncAddMessage(s.contact, s.message, s.attachedContact)
}

type ScenarioCreateGroupAndInvite struct {
	EmptyScenario
	sharedSecret     device.SharedSecret
	groupContact     device.ContactID
	contactsToInvite []device.ContactID
	inviteMessage    string
}

func CreateGroupAndInviteScenario(sharedSecret []byte, contactsToInvite []device.ContactID, inviteMessage string) *ScenarioCreateGroupAndInvite {
	return &ScenarioCreateGroupAndInvite{
		contactsToInvite: contactsToInvite,
		inviteMessage:    inviteMessage,
		sharedSecret:     sharedSecret,
	}
}

func (s *ScenarioCreateGroupAndInvite) OnStart(node *Node) {
	contact, err := node.proto.JoinGroup(s.sharedSecret)
	if err != nil {
		panic(err)
	}

	s.groupContact = contact

	for _, c := range s.contactsToInvite {
		node.proto.SyncAddMessage(c, []byte(s.inviteMessage), &s.groupContact)
	}
}

type ScenarioJoinGroup struct {
	EmptyScenario
	contact     device.ContactID
	groupSecret *device.SharedSecret
	joinedGroup bool
}

func JoinGroupScenario(contact device.ContactID) *ScenarioJoinGroup {
	return &ScenarioJoinGroup{
		contact:     contact,
		groupSecret: nil,
		joinedGroup: false,
	}
}

func (s *ScenarioJoinGroup) Filter(stateUpdate sync.Model) bool {
	for _, node := range stateUpdate.NodeStates {
		for _, message := range node {
			groupSecret := message.AttachedSecret
			if len(groupSecret) != 0 {
				s.groupSecret = &groupSecret
				return true
			}
			return false
		}
	}
	return false
}

func (s *ScenarioJoinGroup) OnStart(node *Node) {
	if s.groupSecret == nil {
		panic("Attempting to join nil group")
	}
	_, err := node.proto.JoinGroup(*s.groupSecret)
	if err != nil {
		panic(err)
	}
	s.joinedGroup = true
}

func (s *ScenarioJoinGroup) JoinedGroup() bool {
	return s.joinedGroup
}

func (s *ScenarioJoinGroup) GroupSecret() *device.SharedSecret {
	return s.groupSecret
}

type ScenarioStoreSyncState struct {
	EmptyScenario
	contact   device.ContactID
	syncState *sync.Model
	syncGoal  SyncGoal
	nodeID    int
}

type SyncGoal struct {
	GroupID           int
	MessagesSent      int
	MessagesDelivered int
	GroupSize         int
}

func NewSyncGoal(groupID int, messagesSent int, messagesDelivered int, groupSize int) SyncGoal {
	return SyncGoal{
		GroupID:           groupID,
		MessagesSent:      messagesSent,
		MessagesDelivered: messagesDelivered,
		GroupSize:         groupSize,
	}
}

func StoreSyncStateScenario(contact device.ContactID, syncGoal SyncGoal, nodeID int) *ScenarioStoreSyncState {
	return &ScenarioStoreSyncState{
		contact:   contact,
		syncState: nil,
		syncGoal:  syncGoal,
		nodeID:    nodeID,
	}
}

func (s *ScenarioStoreSyncState) OnSyncStateChanged(node *Node, contact device.ContactID, stateUpdate sync.Model) {
	if contact == s.contact {
		s.syncState = &stateUpdate
	}
}

func (s *ScenarioStoreSyncState) GetSyncState() *sync.Model {
	return s.syncState
}

func (s *ScenarioStoreSyncState) GetSyncGoal() SyncGoal {
	return s.syncGoal
}

func (s *ScenarioStoreSyncState) GetID() int {
	return s.nodeID
}

func FilterMessageContent(matchMessage string) func(stateUpdate sync.Model) bool {
	return func(stateUpdate sync.Model) bool {
		for _, node := range stateUpdate.NodeStates {
			for _, message := range node {
				return matchMessage == string(message.Value)
			}
		}
		return false
	}
}
