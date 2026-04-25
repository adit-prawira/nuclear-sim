package reactor

import "math"

/*
Coolant temperature rises along with thermal power output
	- Temperature increase, then void fraction increase creating steam bubbles.
	- The void coefficient of the RBMK model indicates that steam drives the value of k-eff to increase, causing a feedback loop
		- steam drives k-eff to increase -> drives more heat -> drives more steam ---
			^																																					|
			|																																					|
			|																																					|
			---------------------------------------------------------------------------
The physics of the coolant phenomenon sugges that:
	- Core temperature rises proportionally to power output minues hear carried away by coolant flow
	- Void fraction increase as coolant temperature approach and exceeds boiling point
		- Boiling point is at ~285°C
	- RBMK void coefficient => +0.005 Δk/1% increase in void fraction. 
		Finetuned to +0.002 Δk/1% increase in void fraction for simulation purposes
	- Cooland behaviour:
		- Reduced coolant flow -> less heat removed -> higher temperature -> more voids
		- Coolant is loss if flow rate is 0 and temperature rises are unchecked
*/									

const (
	// RBMK model operates with nominal flow rate of 8000.0 m^3/h 
	NominalFlowRate = 8000.0
	
	// Coolant Boiling point is at ~285°C 
	BoilingPointC = 285

	// Minimum temperature, preventing coolant going to cold 
	// avoiding reactor shutdown
	MinimumTemperatureC = 265

	// 150% of Nominal flow rate
	MaxFlowRate = 12000.0 
	MinFlowRate = 0.0 

	// RBMK void coefficient => +0.005 Δk/1% increase in void fraction
	// Funetuned for simulation purposes to +0.002 Δk/1% increase in void fraction
	VoidCoefficientPerPercent = 0.002  // was 0.005 (reduce 2.5x)

	MaxNominalVoidFraction = 0.032
)


func (r *reactor) UpdateCoolantFlowRate(delta float64) float64 {
	newFlowRate := r.coolant.FlowRate + delta

	// Upper bound flow rate
	newFlowRate = math.Min(newFlowRate, MaxFlowRate)
	
	// Lower bound flow rate
	newFlowRate = math.Max(newFlowRate, MinFlowRate)

	r.coolant.FlowRate = newFlowRate
	return newFlowRate
}

func (r *reactor) FlowRate() float64 {
	return r.coolant.FlowRate
}

func (r *reactor) VoidFraction() float64 {
	return r.coolant.VoidFraction
}

func (r *reactor) VoidFractionPercent() float64 {
	return r.VoidFraction() * 100
}

func (r *reactor) SetVoidFraction(fraction float64) {
	r.coolant.VoidFraction = fraction
}

func (r *reactor) VoidReactivity() float64 {
	return (r.coolant.VoidFraction*100) * VoidCoefficientPerPercent
}

func (r *reactor) CoolantTemperatureC() float64 {
	return r.coolant.TempC
}

func (r *reactor) SetCoolantTemperatureC(temperature float64) {
	r.coolant.TempC = temperature
}

func (r *reactor) CoolantPressure() float64 {
	return r.coolant.Pressure
}
