package movement_profiles

import (
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
)

// A node in the Random Waypoint Model as defined in "A Performance Comparison of Multi-Hop Wireless Ad Hoc NeWork Routing Protocols": https://dl.acm.org/doi/10.1145/288235.288256
type RandomWaypointNode struct {
	xRange    int
	yRange    int
	maxSpeed  int
	pauseTime time.Duration
	isPaused  bool
	random    *rand.Rand
}

func NewRandomWaypointNode(xRange int, yRange int, maxSpeed int, pauseTime time.Duration, random *rand.Rand) *RandomWaypointNode {
	return &RandomWaypointNode{
		xRange,
		yRange,
		maxSpeed,
		pauseTime,
		false,
		random,
	}
}

func DefaultRandomWaypointNode(random *rand.Rand) *RandomWaypointNode {
	return &RandomWaypointNode{
		1500,
		300,
		20,
		30 * time.Second,
		false,
		random,
	}
}

func (m *RandomWaypointNode) StartPosition() simulator.Coordinate {
	return simulator.Coordinate{
		X: float64(m.random.Intn(m.xRange*2) - m.xRange),
		Y: float64(m.random.Intn(m.yRange*2) - m.yRange),
	}
}

func (m *RandomWaypointNode) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	if !m.isPaused {
		m.isPaused = true
		return simulator.MovementInstruction{
			Coords: coords,
			Time:   m.pauseTime * time.Second,
		}
	} else {
		newCoord := simulator.Coordinate{
			X: float64(m.random.Intn(m.xRange*2) - m.xRange),
			Y: float64(m.random.Intn(m.yRange*2) - m.yRange),
		}
		dist := coords.Distance(newCoord)
		speed := m.random.Float64() * 20

		travelTime := dist * speed

		return simulator.MovementInstruction{
			Coords: newCoord,
			Time:   time.Duration(travelTime) * time.Second,
		}
	}
}
