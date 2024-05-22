package vadere

import (
	"time"

	"github.com/starling-protocol/simulator"
)

type MovementProfile struct {
	coordList  []VadereEntry
	coordIndex int
}

func NewMovementProfile(coordList []VadereEntry) *MovementProfile {
	if len(coordList) < 1 {
		panic("Empty list of coordinates")
	}
	return &MovementProfile{
		coordList:  coordList,
		coordIndex: 0,
	}
}

func (m *MovementProfile) StartPosition() simulator.Coordinate {
	if m.coordIndex != 0 {
		panic("StartPosition called twice on same node")
	}
	return m.coordList[0].Coords
}

func (m *MovementProfile) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	if m.coordIndex >= len(m.coordList) {
		entry := m.coordList[m.coordIndex-1]
		return simulator.MovementInstruction{
			Coords: entry.Coords,
			Time:   time.Duration(1000) * time.Hour,
		}
	}

	entry := m.coordList[m.coordIndex]

	newCoord := entry.Coords
	m.coordIndex++

	return simulator.MovementInstruction{
		Coords: newCoord,
		Time:   entry.Time,
	}
}
