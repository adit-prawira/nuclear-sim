package engine

import (
	"fmt"
	"time"

	"github.com/adit-prawira/nuclear-sim/reactor"
)

type Engine interface {
	Run()
	Stop()
	Tick()
	SimTime() float64
}

type RBMKEngine struct {
	reactor            *reactor.Reactor
	tickInterval       time.Duration
	simulationTickSize float64
	stop               chan struct{}
}

func NewRBMKEnginer(r *reactor.Reactor) Engine {
	return &RBMKEngine{
		reactor:            r,
		tickInterval:       100 * time.Millisecond,
		simulationTickSize: 1.0,
		stop:               make(chan struct{}),
	}
}

// Run implements Engine.
func (r *RBMKEngine) Run() {
	ticker := time.NewTicker(r.tickInterval)

	// call stop when Run() finish running which is triggered when reactor is destroyed
	defer r.Stop() 

	for range ticker.C {
		// Stop when reactor is destroyed due to meltdown
		if r.reactor.SimulationState.Destroyed {
			return
		}
		r.Tick()
		r.print()
	}
}

// SimTime implements Engine.
func (r *RBMKEngine) SimTime() float64 {
	return r.reactor.SimulationState.SimulationTimeSeconds
}

// Stop implements Engine.
func (r *RBMKEngine) Stop() {
	close(r.stop)
}

// Tick implements Engine.
func (r *RBMKEngine) Tick() {
	r.reactor.SimulationState.SimulationTimeSeconds += r.simulationTickSize
}

func (r *RBMKEngine) print() {
	reactor := r.reactor
	fmt.Printf("[t = %4.0fs] Power: %7.1f MW  k-eff: %.3f  Core: %.0f°C  Rods: %d/211  Status: %s\n", 
		reactor.SimulationState.SimulationTimeSeconds,
		reactor.Neutronics.PowerMW,
		reactor.Neutronics.KEffective,
		reactor.Temperature.CoreTempC,
		reactor.ControlRod.RodsInserted,
		reactor.SimulationState.Status.String(),
	)
}


