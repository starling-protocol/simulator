package loggers

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/starling-protocol/simulator"
)

type ProfileLogger struct{}

func NewProfileLogger() *ProfileLogger {
	return &ProfileLogger{}
}

func (l *ProfileLogger) Init() {
	f, err := os.Create("profile.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
}

func (l *ProfileLogger) NewEvent(e simulator.Event) {

	switch e.EventType() {
	case simulator.TERMINATE:
		pprof.StopCPUProfile()
	default:
		return
	}
}

func (l *ProfileLogger) Log(str string) {}
