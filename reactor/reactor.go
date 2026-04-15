package reactor

type ReactorStatus int 

const (
	StatusStable ReactorStatus= iota
	StatusWarning
	StatusSCRAM // this is the status that will indicate when AZ-5 must be used for emergency to avoid meltdown 
	StatusMeltdown
)

func (rs ReactorStatus) String() string {
	switch rs {
	case StatusStable:
		return "STABLE"
	case StatusWarning:
		return "WARNING"
	case StatusSCRAM:
		return "SCRAM"
	case StatusMeltdown:
		return "MELTDOWN"
	default:
		return "UNKNOWN"
	}
}

type Neutronics struct {
	PowerMW float64 // thermal power in megawatt
	KEffective float64 // critically factor 
}

type ControlRod struct {
	RodsInserted int // number of control rods inserted to the core 
}

type Coolant struct {
	FlowRate float64 // Water flow rate in m³/h
	TempC float64 // Water temperature in °C
	VoidFraction float64 // Coolant's fraction (Steam)
	Pressure float64 // Coolant's circuit pressure in bar 
}

type Temperature struct {
	CoreTempC float64 // Reactor's core temperature in °C
	FuelRodTempC float64 // Reactor's fuel rod temperature in °C
}

type XenonPoisoning struct {
	XenonLevel float64 // Xenon-135 concentration coefficient from [0.0, 1.0]
	XenonReactivity float64 // Reactivity penalty in pcm and always negative number
}

type SimulationState struct {
	SimulationTimeSeconds float64 // elapsed simulation time in seconds
	Status ReactorStatus // Reactor 4 different statuses (state machine)
	Destroyed bool // a flag to indicate if reactor is destroyed
}

type Reactor struct {
	Neutronics Neutronics
	ControlRod ControlRod
	Coolant Coolant
	Temperature Temperature 
	XenonPoisoning XenonPoisoning 
	SimulationState SimulationState
}

// Intended normal operational condition for RBMK (Reaktor Bolshoy Moshchnosti Kanalnyy) reactor model
// Russian for High-Power Channel-type Reactor and this is the design used at Chernobyl
func NominalRBMK() *Reactor {
	return &Reactor {
		Neutronics:      Neutronics{
			PowerMW:    1600,
			KEffective: 1.005,
		},
		ControlRod:      ControlRod{
			RodsInserted: 15,
		},
		Coolant:         Coolant{
			FlowRate:     8000,
			TempC:        265,
			VoidFraction: 0.032,
			Pressure:     65,
		},
		Temperature:     Temperature{
			CoreTempC:    285
			FuelRodTempC: 620,
		},
		XenonPoisoning:  XenonPoisoning{
			XenonLevel:      0.124,
			XenonReactivity: -180,
		},
		SimulationState: SimulationState{
			SimulationTimeSeconds: 0,
			Status:                StatusStable,
			Destroyed:             false,
		},
	}
}
