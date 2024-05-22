package transmission_behavior

import (
	"math"
	"time"

	"github.com/starling-protocol/simulator"
)

type BleTransmission struct {
}

func (t BleTransmission) Transmission(originCoord simulator.Coordinate, targetCoord simulator.Coordinate, packet []byte) (shouldBeDropped bool, delay time.Duration) {
	dist := distance(originCoord, targetCoord)
	pathloss := 40 + 25*math.Log10(dist) // 78
	if pathloss > 78 {
		return true, -1
	} else {
		//transmission_speed := float64(len(packet)) / 24_000_000.0 // 24 Mbps max transmission speed for BLE
		return false, time.Duration(2 * time.Millisecond)
	}
}

func distance(coordA simulator.Coordinate, coordB simulator.Coordinate) float64 {
	return math.Sqrt(math.Pow(math.Abs(float64(coordA.X)-float64(coordB.X)), 2) + math.Pow(math.Abs(float64(coordA.Y)-float64(coordB.Y)), 2))
}
