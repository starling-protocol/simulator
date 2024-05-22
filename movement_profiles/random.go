package movement_profiles

import (
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
)

type RandomNode struct {
	x_range  int
	y_range  int
	max_time int
	min_time int
	random   *rand.Rand
}

func NewRandomNode(x_range int, y_range int, max_time int, min_time int, random *rand.Rand) *RandomNode {
	if min_time >= max_time {
		panic("Error creating random movement profile. min_time cannot be less than max_time")
	}
	return &RandomNode{
		x_range,
		y_range,
		max_time,
		min_time,
		random,
	}
}
func (m *RandomNode) StartPosition() simulator.Coordinate {
	return simulator.Coordinate{
		X: float64(m.random.Intn(m.x_range*2) - m.x_range),
		Y: float64(m.random.Intn(m.y_range*2) - m.y_range),
	}
}

func (m *RandomNode) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	newCoord := simulator.Coordinate{
		X: float64(m.random.Intn(m.x_range*2) - m.x_range),
		Y: float64(m.random.Intn(m.y_range*2) - m.y_range),
	}

	if m.max_time-m.min_time == 0 {
		return simulator.MovementInstruction{
			Coords: newCoord,
			Time:   time.Duration(0) * time.Second,
		}
	} else {
		return simulator.MovementInstruction{
			Coords: newCoord,
			Time:   time.Duration(m.random.Intn(m.max_time-m.min_time)+m.min_time) * time.Second,
		}
	}
}
