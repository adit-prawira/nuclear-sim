package physics

import (
	"math"

	"github.com/adit-prawira/nuclear-sim/reactor"
)

type ThermodynamicsEngine interface {
	/*
		A method to update coolant temperature and core temperature
		Values will change depending on current thermal power and coolant flow rate
	*/
	UpdateTemperatures(powerMW, coreTemperature, coolantTemperature, fuelTemperature, flowRate float64, dt float64) (newCoreTemperature, newCoolantTemperature, newFuelTemperature float64)

	/*
		Calculating the void fration of the coolant
		based of the provided void fraction, temperature and pressue
	*/
	CalculateVoidFraction(coolantTemperature, pressue float64) float64	
}

type rbmkThermodynamicsEngine struct {
	boilingPointC float64

	// This is the simplified model of RBMK heat capacity
	// Holding values in J/(kg·K)
	heatCapacity float64

	// Efficiency value determining of how the coolant
	// efficiently removes heat, holding values in W/(m²·K)
	coolingEfficiency float64
}

func NewRBMKThermodynamicsEngine() ThermodynamicsEngine {
	return &rbmkThermodynamicsEngine{
		boilingPointC:      reactor.BoilingPointC,
		// This is a simplified constant
		// as complex physics calculation
		heatCapacity:      0.0001,

		// Calibrated values for the purpose of the simulation 
		// the calculation will involve complex physics calculation
		coolingEfficiency: 0.00002,
	}
}

// CalculateVoidFraction implements ThermodynamicsEngine.
func (rte *rbmkThermodynamicsEngine) CalculateVoidFraction(coolantTemperature float64, pressue float64) float64 {
	// Handle void fraction below boiling point 
	if coolantTemperature < rte.boilingPointC {
		// Linear increases from 0 - 3.2% when approaching boilingPoint 
		fraction := (coolantTemperature - reactor.MinimumTemperatureC) / (rte.boilingPointC - reactor.MinimumTemperatureC)
		return math.Max(0.0, fraction * reactor.MaxNominalVoidFraction)
	}

	temperatureAboveBoilingPoint := coolantTemperature - rte.boilingPointC
	// Exponential rise in void fraction when coolant temperature is above boiling point
	voidFraction := reactor.MaxNominalVoidFraction + (temperatureAboveBoilingPoint * 0.01)

	// upper bound void fraction to be at 100%
	// but if reaches around 80% void fraction means there are no water just steam 
	// heats will increase, causing feedback loop, power keeps increasing 
	// and the reactor is fucked
	return math.Min(voidFraction, 1.0)
}

// UpdateTemperatures implements ThermodynamicsEngine.
func (rte *rbmkThermodynamicsEngine) UpdateTemperatures(powerMW, coreTemperature, coolantTemperature, fuelTemperature, flowRate float64, dt float64) (newCoreTemperature float64, newCoolantTemperature, newFuelTemperature float64) {	
	// Power heats fuel directly
	fuelHeatIn := powerMW * dt * rte.heatCapacity * 5.0 
	coreHeatIn := powerMW * dt * rte.heatCapacity

	// Heat transfer fuel -> core (proportional to temperature difference)
	fuelToCoreTransfer := (fuelTemperature - coreTemperature) * 0.0005 * dt 

	newFuelTemp := fuelTemperature + fuelHeatIn - fuelToCoreTransfer

	heatIn := coreHeatIn + fuelToCoreTransfer

	deltaTemperature := coreTemperature - coolantTemperature
	heatOut := flowRate * rte.coolingEfficiency * deltaTemperature * dt 

	netHeat := heatIn - heatOut
	newCoreTemp := coreTemperature + netHeat 
	
	var newCoolantTemp float64 	
	if flowRate > 0 {
		// coolant ambsorbing heat 
		// Simplified absorbtion mechanics where coolant absrob 30% of heat out 
		heatAbsorbed := heatOut * 0.3 
		newCoolantTemp = coolantTemperature + heatAbsorbed
	} else {
		// Zero flow: coolant heats toward core temperature 
		// Stagnant water heats by radiation/conduction
		heatTransfer := (coreTemperature - coolantTemperature) * 0.01 * dt
		newCoolantTemp = coolantTemperature + heatTransfer
	}
	
	// upper bound coolant temperature with mininum temperature 
	// avoiding coolant from being too cold
	newCoolantTemp = math.Max(newCoolantTemp, reactor.MinimumTemperatureC)

	return newCoreTemp, newCoolantTemp, newFuelTemp
}

