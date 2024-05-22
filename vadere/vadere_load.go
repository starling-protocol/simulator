package vadere

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/starling/utils"
)

type VadereNode struct {
	CoordList []VadereEntry
	StartTime *time.Duration
}

type VadereEntry struct {
	Coords simulator.Coordinate
	Time   time.Duration
}

func VadereLoad(path string, keepRate float64, random *rand.Rand) map[int64]*VadereNode {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ' '
	data, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	data = data[1:]

	coordListMap := map[int64]*VadereNode{}

	for _, line := range data {
		id, err := strconv.Atoi(line[0])
		if err != nil {
			panic("Format error in .traj file")
		}

		coordListMap[int64(id)] = &VadereNode{
			CoordList: []VadereEntry{},
			StartTime: nil,
		}
	}

	for _, vadereNodeId := range utils.ShuffleMapKeys(random, coordListMap) {
		r := random.Float64()
		if r > keepRate {
			delete(coordListMap, vadereNodeId)
		}
	}

	for _, line := range data {
		id, err := strconv.Atoi(line[0])
		if err != nil {
			panic("Format error in .traj file")
		}
		_, found := coordListMap[int64(id)]
		if !found {
			continue
		}

		startTime, err2 := strconv.ParseFloat(line[1], 32)
		endTime, err3 := strconv.ParseFloat(line[2], 32)
		x, err4 := strconv.ParseFloat(line[5], 32)
		y, err5 := strconv.ParseFloat(line[6], 32)

		if err2 != nil || err3 != nil || err4 != nil || err5 != nil {
			panic("Format error in .traj file")
		}

		newEntry := VadereEntry{
			Coords: simulator.Coordinate{X: x, Y: y},
			Time:   time.Duration((endTime - startTime) * 1_000_000_000),
		}
		coordListMap[int64(id)].CoordList = append(coordListMap[int64(id)].CoordList, newEntry)
		if coordListMap[int64(id)].StartTime == nil {
			delay := time.Duration(startTime) * 1_000_000_000
			coordListMap[int64(id)].StartTime = &delay
		}
	}

	return coordListMap
}
