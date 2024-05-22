package movement_profiles

import (
	"time"

	"github.com/starling-protocol/simulator"
)

type StationaryNode struct {
	StartCoords simulator.Coordinate
}

func NewStationary(x float64, y float64) *StationaryNode {
	return &StationaryNode{
		StartCoords: simulator.Coordinate{X: x, Y: y},
	}
}

func (m *StationaryNode) StartPosition() simulator.Coordinate {
	return m.StartCoords
}

func (m *StationaryNode) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	return simulator.MovementInstruction{
		Coords: coords,
		Time:   1 * time.Second,
	}
}
