package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/adit-prawira/nuclear-sim/physics"
	"github.com/adit-prawira/nuclear-sim/reactor"
	"golang.org/x/term"
)

type Engine interface {
	Run()	
}

type RBMKEngine struct {
	reactor            reactor.Reactor
	tickInterval       time.Duration
	simulationTickSize float64
	neutronics         physics.NeutronicsEngine
	thermodynamics     physics.ThermodynamicsEngine
	stopChan           chan struct{}
	inputChan 				 chan byte 
	graphiteSpikeActive bool 
}

func NewRBMKEnginer(r reactor.Reactor) Engine {
	return &RBMKEngine{
		reactor:            r,
		tickInterval:       100 * time.Millisecond,
		neutronics:         physics.NewRBMKNeutronicsEngine(),
		thermodynamics:     physics.NewRBMKThermodynamicsEngine(),
		simulationTickSize: 1.0,
		stopChan:           make(chan struct{}),
		inputChan: 					make(chan byte, 10),
		graphiteSpikeActive: false,
	}
}

// Run implements Engine.
func (re *RBMKEngine) Run() {
	ticker := time.NewTicker(re.tickInterval)
	// set terminal raw mode 
	currentState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("ERROR: Unable to set terminal to raw mode: %v\n", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), currentState)

	// Listen for keyboard input 
	go re.listenForInput()

	// call stop when Run() finish running which is triggered when reactor is destroyed
	defer re.stop() 

	re.printControls()

	for range ticker.C {
		// Process keyboard input 
		re.processInput()

		// Stop when reactor is destroyed due to meltdown
		if re.reactor.IsDestroyed() {
			re.printMeltdown()
			return
		}
		re.tick()
		re.print()
	} 
}

func (re *RBMKEngine) stop() {
	close(re.stopChan)
}

// Tick implements Engine.
func (re *RBMKEngine) tick() {
	dt := re.simulationTickSize
	
	baseKeff := 1.000  
	rodReactivity := re.reactor.RodReactivity()
	voidReactivity := re.reactor.VoidReactivity()
	xenonReactivity := re.reactor.XenonReactivity()
	graphiteSpike := 0.0 
	
	if re.graphiteSpikeActive {
		graphiteSpike = reactor.GraphiteTipSpike // +0.0003
		re.graphiteSpikeActive = false
	}

	re.reactor.SetKEffective(baseKeff + rodReactivity + voidReactivity + xenonReactivity + graphiteSpike)

	currentPower := re.reactor.ThermalPower()
	keff := re.reactor.KEffective()
	re.reactor.SetThermalPower(re.neutronics.UpdatePower(currentPower, keff, dt))
	re.reactor.UpdateXenon(dt)

	newCoreTemperature, newCoolantTemperature := re.thermodynamics.UpdateTemperatures(
		re.reactor.ThermalPower(),
		re.reactor.CoreTemperatureC(),
		re.reactor.CoolantTemperatureC(),
		re.reactor.FlowRate(),
		dt,
	)
	re.reactor.SetCoreTemperatureC(newCoreTemperature)
	re.reactor.SetCoolantTemperatureC(newCoolantTemperature)

	voidFraction := re.thermodynamics.CalculateVoidFraction( 
		re.reactor.CoolantTemperatureC(), 
		re.reactor.CoolantPressure(),
	)
	re.reactor.SetVoidFraction(voidFraction)

	re.reactor.UpdateSimulationTimeSeconds(dt)		
	re.updateStatus()
}

func (re *RBMKEngine) print() {
	reactor := re.reactor
	fmt.Printf("\r[t = %4.0fs] Power: %7.1f MW  k-eff: %.3f  Core: %.0f°C  Void: %4.1f%%	 Xenon: %4.1f%%  Flow: %.0f m^3/h  Rods: %d/211  Status: %s\r\n", 
		reactor.SimulationTimeSeconds(),
		reactor.ThermalPower(),
		reactor.KEffective(),
		reactor.CoreTemperatureC(),
		reactor.VoidFractionPercent(),
		reactor.XenonLevelPercent(),
		reactor.FlowRate(),
		reactor.TotalInsertedRods(),
		reactor.Status().String(),
	)
}

func (re *RBMKEngine) updateStatus() {	
	switch {
	case re.reactor.IsMeltdown():
		re.reactor.SetStatus(reactor.StatusMeltdown)
		re.reactor.SetIsDestroyed(true)		
	case re.reactor.IsSupercritical():
		re.reactor.SetStatus(reactor.StatusWarning) 
	default:
		re.reactor.SetStatus(reactor.StatusStable) 
	}
}

func (re *RBMKEngine) printMeltdown() {
	fmt.Printf("\r\n")
	fmt.Printf("\r!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\r\n")
	fmt.Printf("   REACTOR DESTROYED\r\n")
	fmt.Printf("   1986-04-26  %s\r\n", re.simTimeAsClockString())
	fmt.Printf("   PROMPT CRITICALITY EXCEEDED\r\n")
	fmt.Printf("   Peak power: %.0f MW  (%.0f%% of nominal)\r\n",
					re.reactor.ThermalPower(),
					re.reactor.PowerPecentOfNominal(),
	)
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\r\n")
}

func (re *RBMKEngine) simTimeAsClockString() string {
	totalSeconds := int(re.reactor.SimulationTimeSeconds())
	baseHour := 1
	baseMinute := 22
	baseSecond := 30

	totalSeconds += baseHour*3600 + baseMinute*60 + baseSecond
	h := (totalSeconds / 3600) % 24
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (re *RBMKEngine) printControls() {
	fmt.Println("\r\nControls:\r")
	fmt.Println("		i - Insert 1 control rod\r")
	fmt.Println("		I - Insert 10 control rods\r")
	fmt.Println("		o - Withdraw 1 control rod\r")
	fmt.Println("		O - Withdraw 10 control rods\r")
	fmt.Println("		f - Increase coolant flow rate\r")
	fmt.Println("		F - Decrease coolant flow rate\r")
	fmt.Println("		q - Quit\r\n\r")
	fmt.Println()
}

func (re *RBMKEngine) listenForInput() {
	buf := make([]byte, 1)
	for {
		select {
		case <- re.stopChan:
			return
		default:
			n, err := os.Stdin.Read(buf)
			
			// Continue to stream for input
			if err != nil || n == 0 {
				continue
			} 

			// Otherwise set input
			re.inputChan <- buf[0] 
		}
	}
}

func (re * RBMKEngine	) processInput() {
	for {
		select {
		case input := <- re.inputChan: 
			switch input {
			case 'i': 
				inserted, isGraphiteTipSpike := re.reactor.InsertRods(1)
				re.logControlRodsInsertion(inserted, isGraphiteTipSpike)
			case 'I': 
				inserted, isGraphiteTipSpike := re.reactor.InsertRods(10)
				re.logControlRodsInsertion(inserted, isGraphiteTipSpike)
			case 'o': 
				withdrawn, isBelowSafe := re.reactor.WithdrawnRods(1)
				re.logControlRodsWithdrawal(withdrawn, isBelowSafe)
			case 'O': 
				withdrawn, isBelowSafe := re.reactor.WithdrawnRods(10)
				re.logControlRodsWithdrawal(withdrawn, isBelowSafe)
			case 'f': 
				newFlowRate := re.reactor.UpdateCoolantFlowRate(1000.0)
				fmt.Printf("\r> Increase coolant flow to %.0f m^3/h\r\n", newFlowRate)
			case 'F': 
				newFlowRate := re.reactor.UpdateCoolantFlowRate(-1000.0)
				fmt.Printf("\r> Decrease coolant flow to %.0f m^3/h\r\n", newFlowRate)
			case 'q', 'Q': 
        fmt.Println("\r\n> System Shutdown Requested")			  
				re.reactor.SetIsDestroyed(true)
			}	
		default: 
			return
		}
	}
}

func (re *RBMKEngine) logControlRodsInsertion(inserted int, isGraphiteTipSpike bool) {
	if isGraphiteTipSpike {
		re.graphiteSpikeActive = true 
		fmt.Printf("\r> Inserted %d control rod (GRAPHITE TIP SPIKE!)\r\n", inserted)
	}else {
		fmt.Printf("\r> Inserted %d conrol rod\r\n", inserted)
	}
}

func (re *RBMKEngine) logControlRodsWithdrawal(withdrawn int, isBelowSafe bool) {
	fmt.Printf("\r> Withdrew %d control rod\r\n", withdrawn)
	if isBelowSafe{
		fmt.Printf("\rWARNING: %d control rod(s) inserted - below safe minium of %d control rods\r\n", re.reactor.TotalInsertedRods(), reactor.MinimumSafeRods)
	}
}

