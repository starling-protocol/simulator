package transmission_behavior

import (
	"time"

	"github.com/starling-protocol/simulator"
)

type NoDrops struct{}

func (t NoDrops) Transmission(originCoord simulator.Coordinate, targetCoord simulator.Coordinate, packet []byte) (shouldBeDropped bool, delay time.Duration) {
	return false, time.Duration(0)
}
