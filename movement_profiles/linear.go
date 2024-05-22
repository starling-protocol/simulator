package movement_profiles

import (
	"time"

	"github.com/starling-protocol/simulator"
)

type LinearNode struct {
	StartCoords  simulator.Coordinate
	EndCoords    simulator.Coordinate
	Seconds      time.Duration
	reachedStart bool
}

func NewLinearNode(startCoords simulator.Coordinate, endCoords simulator.Coordinate, seconds int) *LinearNode {
	return &LinearNode{
		StartCoords:  startCoords,
		EndCoords:    endCoords,
		Seconds:      time.Duration(seconds) * time.Second,
		reachedStart: true,
	}
}

func (m *LinearNode) StartPosition() simulator.Coordinate {
	return m.StartCoords
}

func (m *LinearNode) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	if m.reachedStart {
		m.reachedStart = false
		return simulator.MovementInstruction{m.EndCoords, m.Seconds}
	} else {
		m.reachedStart = true
		return simulator.MovementInstruction{m.StartCoords, m.Seconds}
	}
}
