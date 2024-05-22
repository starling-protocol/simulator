package transmission_behavior

import (
	"time"

	"github.com/starling-protocol/simulator"
)

type DelayTransmission struct {
	Delay time.Duration
}

func (t DelayTransmission) Transmission(originCoord simulator.Coordinate, targetCoord simulator.Coordinate, packet []byte) (shouldBeDropped bool, delay time.Duration) {
	return false, t.Delay
}
