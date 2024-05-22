package transmission_behavior

import (
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
)

type RandomDrops struct {
	DropChance float64
	Delay      time.Duration
	Random     *rand.Rand
}

func (t RandomDrops) Transmission(originCoord simulator.Coordinate, targetCoord simulator.Coordinate, packet []byte) (shouldBeDropped bool, delay time.Duration) {
	if t.Random.Float64() < t.DropChance {
		return true, t.Delay
	} else {
		return false, t.Delay
	}
}
