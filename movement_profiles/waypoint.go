package movement_profiles

import (
	"github.com/starling-protocol/simulator"
)

type WaypointNode struct {
	points []simulator.MovementInstruction
}

func NewWaypointNode() *WaypointNode {
	return &WaypointNode{
		points: []simulator.MovementInstruction{},
	}
}

func (w *WaypointNode) AddPoint(point simulator.MovementInstruction) *WaypointNode {
	w.points = append(w.points, point)
	return w
}

func (w *WaypointNode) StartPosition() simulator.Coordinate {
	return w.points[0].Coords
}

func (w *WaypointNode) RegisterMovements(coords simulator.Coordinate) simulator.MovementInstruction {
	if len(w.points) == 0 {
		panic("no more waypoints")
	}

	point := w.points[0]
	w.points = w.points[1:]
	return point
}
