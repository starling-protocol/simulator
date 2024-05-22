package transmission_behavior

import (
	"math"
	"math/rand"
	"time"

	"github.com/starling-protocol/simulator"
)

type LongTailTransmission struct {
	Alpha      float64 // A float representing the Pareto index of a Pareto distribution. The lower it is, the more likely that transmissions have a large delay
	X_m        float64 // A float representing the x_m in a Pareto distribution. This is the minimum time it can take to transmit a packet
	DropChance float64 // The chance that a packet is dropped
	Random     rand.Rand
}

func (t LongTailTransmission) Transmission(originCoord simulator.Coordinate, targetCoord simulator.Coordinate, packet []byte) (shouldBeDropped bool, delay time.Duration) {

	if t.Random.Float64() < t.DropChance {
		return true, -1
	} else {

		rnd := rand.ExpFloat64()
		delay := t.X_m * math.Exp(rnd/t.Alpha)
		return false, time.Duration(time.Duration(delay) * time.Millisecond)
	}
}
