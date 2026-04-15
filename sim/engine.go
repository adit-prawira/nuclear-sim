package engine

import (
	"fmt"
	"time"

	"github.com/adit-prawira/nuclear-sim/physics"
	"github.com/adit-prawira/nuclear-sim/reactor"
)

type Engine interface {
	Run()
	SimTime() float64
}

type RBMKEngine struct {
	reactor            *reactor.Reactor
	tickInterval       time.Duration
	simulationTickSize float64
	neutronics         physics.NeutronicsEngine
	stopChan           chan struct{}
}

func NewRBMKEnginer(r *reactor.Reactor) Engine {
	return &RBMKEngine{
		reactor:            r,
		tickInterval:       100 * time.Millisecond,
		neutronics:         physics.NewRBMKNeutronicsEngine(),
		simulationTickSize: 1.0,
		stopChan:           make(chan struct{}),
	}
}

// Run implements Engine.
func (re *RBMKEngine) Run() {
	ticker := time.NewTicker(re.tickInterval)

	// call stop when Run() finish running which is triggered when reactor is destroyed
	defer re.stop() 

	for range ticker.C {
		// Stop when reactor is destroyed due to meltdown
		if re.reactor.SimulationState.Destroyed {
			re.printMeltdown()
			return
		}
		re.tick()
		re.print()
	}
}

func (re *RBMKEngine) SimTime() float64 {
	return re.reactor.SimulationState.SimulationTimeSeconds
}

func (re *RBMKEngine) stop() {
	close(re.stopChan)
}

// Tick implements Engine.
func (re *RBMKEngine) tick() {
	dt := re.simulationTickSize
	currentPower := re.reactor.Neutronics.PowerMW
	keff := re.reactor.Neutronics.KEffective

	re.reactor.SimulationState.SimulationTimeSeconds += dt	
	re.reactor.Neutronics.PowerMW = re.neutronics.UpdatePower(currentPower, keff, dt)
	re.updateStatus()
}

func (re *RBMKEngine) print() {
	reactor := re.reactor
	fmt.Printf("[t = %4.0fs] Power: %7.1f MW  k-eff: %.3f  Core: %.0f°C  Rods: %d/211  Status: %s\n", 
		reactor.SimulationState.SimulationTimeSeconds,
		reactor.Neutronics.PowerMW,
		reactor.Neutronics.KEffective,
		reactor.Temperature.CoreTempC,
		reactor.ControlRod.RodsInserted,
		reactor.SimulationState.Status.String(),
	)
}

func (re *RBMKEngine) updateStatus() {
	// Nominal RBMK power is 3200 MW 
	// Historical estimates put the peak power during the explosion at ~30,000 MW  to  ~33,000 MW
	// Therefore, approximately 3200 * 10 ~= 32,000 MW 
	// Use 32,000 MW as meltdown threshold
	switch {
	case re.reactor.Neutronics.PowerMW >= 32000:
		re.reactor.SimulationState.Status = reactor.StatusMeltdown
		re.reactor.SimulationState.Destroyed = true

	// Since 3200 is the max of RBMK nominal thermal power 
	// start warning when goes beyong 
	// or if keff > 1.0 start warning because reaction is growing (supercritical)
	case re.reactor.Neutronics.PowerMW >= 3200 || re.reactor.Neutronics.KEffective > 1.0:
		re.reactor.SimulationState.Status = reactor.StatusWarning
	default:
		re.reactor.SimulationState.Status = reactor.StatusStable
	}
}

func (re *RBMKEngine) printMeltdown() {
	fmt.Printf("\n")
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Printf("   REACTOR DESTROYED\n")
	fmt.Printf("   1986-04-26  %s\n", re.simTimeAsClockString())
	fmt.Printf("   PROMPT CRITICALITY EXCEEDED\n")
	fmt.Printf("   Peak power: %.0f MW  (%.0f%% of nominal)\n",
					re.reactor.Neutronics.PowerMW,
					(re.reactor.Neutronics.PowerMW/3200)*100,
	)
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
}

func (re *RBMKEngine) simTimeAsClockString() string {
	totalSeconds := int(re.reactor.SimulationState.SimulationTimeSeconds)
	baseHour := 1
	baseMinute := 22
	baseSecond := 30

	totalSeconds += baseHour*3600 + baseMinute*60 + baseSecond
	h := (totalSeconds / 3600) % 24
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
