package physics

import "math"

type NeutronicsEngine interface {
	// Calculate thermal power level based of k-effective
	// dP/dt = ((k-eff - 1) / λ) * P
	// => First linear ODE -> P(t + dt) = P(t) * e^(((k-eff - 1 - β) / λ) * dt)
	//									   -> P(t + dt) = P(t) * e^(α * dt)
	// Where:
	// 	- P(t + dt) is the new power
	//  - P(t) is the current power
	//  - α is the reactivity rate coefficient, α = (k-eff - 1 - β ) / λ
	//  - β is the fraction of neutrons that are delayed
	//  - λ is tge prompted neutron lifetime 
	UpdatePower(currentPower, keff float64, dt float64) float64
}

type NeutronicsConfig struct {
	// β -> fraction of neutrons that are delayed
	DelayedNeutronFraction float64
	// λ -> prompted neutron lifetime
	PromptNeutronLifetime float64
}

type rbmkNeutronicsEngine struct {
	config NeutronicsConfig
}

type delayedNeutronGroup struct {
	// fraction of β contributed by this group
	BetaFraction float64
	// half-life in seconds
	HalfLife float64
}

// | Group | Half-life (s)| Fraction of β | Mean Lifetime (s) | Contribution (s) |
// |-------|--------------|---------------|-------------------|------------------|
// | 1     | 55.0         | 0.033         | 79.37             | 2.62             |
// | 2     | 22.0         | 0.219         | 31.75             | 6.95             |
// | 3     | 6.0          | 0.196         | 8.66              | 1.70             |
// | 4     | 2.0          | 0.395         | 2.89              | 1.14             |
// | 5     | 0.5          | 0.115         | 0.72              | 0.08             |
// | 6     | 0.2          | 0.042         | 0.29              | 0.01             |
// | **Σ** |              | **1.000**     |                   | **12.50s**       |
var rbmkDelayedNeutronGroups = []delayedNeutronGroup{
	{BetaFraction: 0.033, HalfLife: 55.0},
	{BetaFraction: 0.219, HalfLife: 22.0},
	{BetaFraction: 0.196, HalfLife: 6.0},
	{BetaFraction: 0.395, HalfLife: 2.0},
	{BetaFraction: 0.115, HalfLife: 0.5},
	{BetaFraction: 0.042, HalfLife: 0.2},
}


func NewRBMKNeutronicsEngine() NeutronicsEngine {
	return &rbmkNeutronicsEngine{
		config: NeutronicsConfig{
			// For RBMK scenario it is 0.0065
			DelayedNeutronFraction: 0.0065,
			// For RBMK scenario it is 0.0001s
			PromptNeutronLifetime: 0.0001,
		},
	}
}


// UpdatePower implements NeutronicsEngine.
func (rne *rbmkNeutronicsEngine) UpdatePower(currentPower float64, keff float64, dt float64) float64 {
	beta := rne.config.DelayedNeutronFraction
	lambda := rne.config.PromptNeutronLifetime

	// Regime 2: Supercritical prompt (k-ef >= 1 + β)
	if keff >= 1 + beta {
		alpha := (keff - 1 - beta) / lambda
		power := currentPower * math.Exp(alpha * dt)
		power = math.Max(power, 0.01)
		return power 
	}

	// Regime 1: Subcritical Prompt (k-eff < 1 + β)
	// λ_eff = β * τ_d 
	lambdaEff := beta * math.Ceil(rne.meanDelayedNeutronLifetime(rbmkDelayedNeutronGroups))
	alpha := (keff - 1.0) / lambdaEff
	power := currentPower * math.Exp(alpha * dt)
	power = math.Max(power, 0.01)	
	return power 
}

// τ_d = Σ (fraction_i * halfLife_i / ln(2))
func (re *rbmkNeutronicsEngine) meanDelayedNeutronLifetime(groups []delayedNeutronGroup) float64 {
	tau := 0.0 
	for _, g := range groups {
		tau += (g.BetaFraction * g.HalfLife)/math.Ln2 
	}
	return tau
}
