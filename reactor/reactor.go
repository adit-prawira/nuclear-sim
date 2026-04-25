package reactor

type ReactorStatus int 

const (
	StatusStable ReactorStatus= iota
	StatusWarning

	// this is the status that will indicate when 
	// AZ-5 must be used for emergency to avoid meltdown 
	StatusSCRAM 
	StatusMeltdown
)

const (
	MeltdownThermalPower = 32000
	NominalThermalPower = 3200
	
	// -0.003 pcm/°C = -0.00000003 Δk/°C (real RBMK value)
	// Gameplay-tuned to -0.0002 Δk/°C for effective negative feedback
	DopplerTemperatureCoefficient = -0.0002 
	NominalFuelTemperatureC = 620.0
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

type Reactor interface {
	IsDestroyed() bool
	SetIsDestroyed(value bool)

	ThermalPower() float64
	SetThermalPower(newPower float64)

	/*
	Meltdown conditions: 
		- 🔥 Fuel >1500°C (melting point)
		- ⚡ Power >32,000 MW (10x nominal)
		- 🌡️ Core >600°C (structural failure)
	*/
	IsFuelTemperatureCMeltdown() bool
	IsCoreTemperatureCMeltdown() bool 
	IsThermalPowerMeltdown() bool 

	IsSupercritical() bool
	PowerPecentOfNominal() float64	

	KEffective() float64
	SetKEffective(keff float64)

	CoreTemperatureC() float64 
	SetCoreTemperatureC(temperature float64)

	Status() ReactorStatus
	SetStatus(status ReactorStatus)

	SimulationTimeSeconds() float64
	UpdateSimulationTimeSeconds(dt float64)
	
	IsMeltdown() bool

	// Control rods
	TotalInsertedRods() int 
	InsertRods(count int) (inserted int, isGraphiteTipSpike bool)
	WithdrawnRods(count int) (withdrawn int, isBelowSafe bool)
	RodReactivity() float64

	// Coolant 
	UpdateCoolantFlowRate(delta float64) float64
	VoidReactivity() float64 
	FlowRate() float64 
	
	VoidFraction() float64
	VoidFractionPercent() float64 
	SetVoidFraction(fraction float64)

	CoolantTemperatureC() float64 
	SetCoolantTemperatureC(temperature float64) 

	CoolantPressure() float64

	// Xenon-135 
	UpdateXenon(dt float64)
	XenonLevel() float64 
	SetXenonLevel(level float64)
	XenonLevelPercent() float64
	XenonReactivity() float64
	SetXenonReactivity(reactivity float64)

	// Doopler Feedback loop
	FuelTemperatureC() float64 
	SetFuelTemperatureC(temperature float64)
	DopplerReactivity() float64
}

type reactor struct {
	neutronics Neutronics
	controlRod ControlRod
	coolant Coolant
	temperature Temperature 
	xenonPoisoning XenonPoisoning 
	simulationState SimulationState
}

// Intended normal operational condition for RBMK (Reaktor Bolshoy Moshchnosti Kanalnyy) reactor model
// Russian for High-Power Channel-type Reactor and this is the design used at Chernobyl
func NominalRBMK() Reactor {
	return &reactor {
		neutronics:      Neutronics{
			PowerMW:    1600,
			KEffective: 1.000, 
		},
		controlRod:      ControlRod{
			RodsInserted: 15,
		},
		coolant:         Coolant{
			FlowRate:     8000,
			TempC:        265,
			VoidFraction: 0.0,
			Pressure:     65,
		},
		temperature:     Temperature{
			CoreTempC:    285,
			FuelRodTempC: 620,
		},
		xenonPoisoning:  XenonPoisoning{
			XenonLevel:      0.0,
			XenonReactivity: 0.0,
		},
		simulationState: SimulationState{
			SimulationTimeSeconds: 0,
			Status:                StatusStable,
			Destroyed:             false,
		},
	}
}

func (r *reactor) IsDestroyed() bool {
	return r.simulationState.Destroyed
}

func (r *reactor) SetIsDestroyed(value bool) {
	r.simulationState.Destroyed = value
}

func (r *reactor) ThermalPower() float64 {
	return r.neutronics.PowerMW
}

func (r *reactor) SetThermalPower(newPower float64) {
	r.neutronics.PowerMW = newPower
}

// Nominal RBMK power is 3200 MW 
// Historical estimates put the peak power during the explosion at ~30,000 MW  to  ~33,000 MW
// Therefore, approximately 3200 * 10 ~= 32,000 MW 
// Use 32,000 MW as meltdown threshold
func (r *reactor) IsMeltdown() bool {
	return r.neutronics.PowerMW >= MeltdownThermalPower
}

// Since 3200 is the max of RBMK nominal thermal power 
// start warning when goes beyong 
// or if keff > 1.0 start warning because reaction is growing (supercritical)
func (r *reactor) IsSupercritical() bool {
	return r.neutronics.PowerMW >= NominalThermalPower || r.neutronics.KEffective >= 1.0
}

func (r *reactor) PowerPecentOfNominal() float64 {
	return (r.neutronics.PowerMW/NominalThermalPower) * 100.0
}

func (r * reactor) KEffective() float64{
	return r.neutronics.KEffective
}

func (r *reactor) SetKEffective(keff float64) {
	r.neutronics.KEffective = keff 
}

func (r *reactor) CoreTemperatureC() float64 {
	return r.temperature.CoreTempC
}

func (r *reactor) SetCoreTemperatureC(temperature float64) {
	r.temperature.CoreTempC = temperature
}

func (r *reactor) Status() ReactorStatus {
	return r.simulationState.Status
}

func (r *reactor) SetStatus(status ReactorStatus) {
	r.simulationState.Status = status
}

func (r *reactor) SimulationTimeSeconds() float64 {
	return r.simulationState.SimulationTimeSeconds
}

func (r *reactor) UpdateSimulationTimeSeconds(dt float64) {
	r.simulationState.SimulationTimeSeconds += dt
}

func (r *reactor) FuelTemperatureC() float64 {
	return r.temperature.FuelRodTempC
}

func (r *reactor) SetFuelTemperatureC(temperature float64) {
	r.temperature.FuelRodTempC = temperature
}

func (r *reactor) DopplerReactivity() float64 {
	deltaTemperature := r.temperature.FuelRodTempC - NominalFuelTemperatureC
	return DopplerTemperatureCoefficient * deltaTemperature
}

func (r *reactor) IsFuelTemperatureCMeltdown() bool {
	return r.FuelTemperatureC() > 1500.0	
}

func (r *reactor) IsThermalPowerMeltdown() bool {
	return r.ThermalPower() > NominalThermalPower * 10
}

func (r *reactor) IsCoreTemperatureCMeltdown() bool {
	return r.CoreTemperatureC() > 600.0 
}
