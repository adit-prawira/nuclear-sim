package reactor

import "math"

/*
Xenon-135 is a fission byproduct and suppresses the chain reaction of neutron atoms colliding.
	- Xenon-135 builds up faster when reactor is at low thermal power
	- At low power xenon builds up faster than it burns, creating xenon
		pit making it difficult to restart or raise power of the reactor.

Physics Rule:
	- Xenon production rate is proportional to power level (fission rate)
	- Xenon burns by neutron flux and natural decay (9.2 hour hour half-life)
	- At full thermal power, xenon burn rate exceeds production - Xenon
		stays low
	- At low thermal power, xenon production exceeds burn rate - xenon
		accumulates

	- Maximum xenon worth is at ~ -2800pcm, where if it goes unchecked
		reactor is down cause critical outage. (Especially the city around
		the RBMK power plan)

	- Control rods must be withdrawn to increase power, reducing xenon build up.
*/

const (
	// Xenon-135 half-life is 9.2 hours ~ 33,120 seconds 
	XenonHalfLifeSeconds = 33120.0

	// Xenon max worth is -2,800 pcm = ~ 0.028 Δk
	MaxXenonWorth = -0.028

	// Values to be fine tuned for simulation
	XenonProductionRate = 0.000001

	// Xenon burn rate coefficient (depends on neutron flux and power)
	XenonBurnRate = 0.000002

	// Xenon decay constant is λ = ln(2) / half-life
	// It take hours for xenon to decay therefore 
	// compress decay constant to 10 minutes for
	// simulation purposes λ ~ 0.001 
	XenonDecayConstant = 0.001 

)

// Calculate and update xenon level based of 
// production, burn, and decary 
func (r *reactor) UpdateXenon(dt float64){
	currentLevel := r.XenonLevel()
	currentPower := r.ThermalPower()

	// xenon production proportional to fission rate 
	production := XenonProductionRate * currentPower * dt

	// xenon burn proportional to neutron flux and half-life decay
	burn := XenonBurnRate * currentPower * currentLevel * dt

	// xenon natural decay 
	decay := XenonDecayConstant * currentLevel * dt

	newXenonLevel := currentLevel + production - burn - decay
	newXenonLevel = math.Max(newXenonLevel, 0.0)
	newXenonLevel = math.Min(newXenonLevel, 1.0)
	
	r.SetXenonLevel(newXenonLevel)

	// Reactivity penalty
	r.SetXenonReactivity(newXenonLevel * MaxXenonWorth)
}

func (r *reactor) XenonLevel() float64 {
	return r.xenonPoisoning.XenonLevel
}

func (r *reactor) SetXenonLevel(level float64) {
	r.xenonPoisoning.XenonLevel = level
}

func (r *reactor) XenonLevelPercent() float64 {
	return r.XenonLevel() * 100
}

func (r *reactor) XenonReactivity() float64 {
	return r.xenonPoisoning.XenonReactivity
}

func (r *reactor) SetXenonReactivity(reactivity float64) {
	r.xenonPoisoning.XenonReactivity = reactivity
}
