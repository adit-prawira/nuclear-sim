# Phase 1 — The Core Engine

The goal of Phase 1 is a reactor that exists, ticks, responds to operator input,
and can go critical and melt down. No graphics yet — pure physics breathing in
the console.

Each vertical slice delivers a runnable program. Never a dead layer waiting on
another.

---

## Vertical Slices

### Slice 1 — A Reactor That Exists and Breathes

**Goal**: The reactor has state. The engine ticks. The console prints each second.

**Files touched**:
- `reactor/reactor.go` — reactor state struct + NominalRBMK() constructor
- `sim/engine.go` — tick loop, print state each tick
- `main.go` — wire it together, start the engine

**Done when**:
- `go run main.go` prints reactor state every 100ms
- State is static but correct — power, k-eff, temperature, rod count, status

**Expected output**:
```
RBMK-1000 Nuclear Reactor Simulator
Chernobyl Nuclear Power Plant — Unit 4
========================================
[t=   1s]  Power:  1600.0 MW  k-eff: 1.000  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   2s]  Power:  1600.0 MW  k-eff: 1.000  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   3s]  Power:  1600.0 MW  k-eff: 1.000  Core: 285°C  Rods: 15/211  Status: STABLE
```

---
### Slice 2 — k-eff Drives Power

**Goal**: Each tick, k-eff calculates a new power level. Power grows when k-eff > 1,
falls when k-eff < 1. The reactor is now alive.

**Files touched**:
- `physics/neutronics.go` — k-eff to power calculation, two-regime point kinetics model
- `sim/engine.go` — call neutronics each tick before printing

**Physics introduced**:

Two-regime point kinetics model based on the relationship between k-eff and the
prompt criticality threshold (1 + β = 1.0065):

Regime 1 — Subcritical Prompt (k-eff < 1 + β):
- Delayed neutrons dominate — reactor is slow and controllable
- Effective lifetime: λ_eff = β * τ_d where τ_d = 13s (mean delayed neutron lifetime)
- α = (k-eff - 1) / λ_eff
- Power changes gradually — observable at human timescales

Regime 2 — Supercritical Prompt (k-eff >= 1 + β):
- Prompt neutrons alone sustain the reaction — extremely fast
- α = (k-eff - 1 - β) / λ where λ = 0.0001s (prompt neutron lifetime)
- Power surges instantly — meltdown within the same tick
- This is the Chernobyl moment

Exact solution for both regimes:
- P(t + dt) = P(t) * e^(α * dt)
- Power clamped to 0 at minimum
- Status: WARNING when power >= 3200 MW or k-eff > 1.0

**Done when**:
- k-eff = 1.000 holds power steady at 1600 MW
- k-eff = 0.990 causes slow power fall
- k-eff = 1.005 causes slow, visible power rise (subcritical prompt)
- k-eff = 1.010 causes immediate meltdown (supercritical prompt — above 1 + β)
- Status correctly shows WARNING when power exceeds nominal or k-eff exceeds 1.0

**Expected output**:

Subcritical prompt (k-eff = 1.005):
```
[t=   1s]  Power:  1600.0 MW  k-eff: 1.000  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   2s]  Power:  1697.3 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   3s]  Power:  1800.6 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   4s]  Power:  1910.3 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   5s]  Power:  2027.0 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   6s]  Power:  2151.3 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   7s]  Power:  2283.7 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   8s]  Power:  2424.8 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=   9s]  Power:  2575.3 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=  10s]  Power:  2736.0 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=  11s]  Power:  2907.5 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=  12s]  Power:  3090.8 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=  13s]  Power:  3286.8 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: STABLE
[t=  14s]  Power:  3496.5 MW  k-eff: 1.005  Core: 285°C  Rods: 15/211  Status: WARNING
```

Supercritical prompt (k-eff = 1.010 — above 1 + β = 1.0065):
```
[t=   1s]  Power: 32000.0 MW  k-eff: 1.010  Core: 285°C  Rods: 15/211  Status: MELTDOWN

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 REACTOR DESTROYED
 1986-04-26  01:23:44
 PROMPT CRITICALITY EXCEEDED
  !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
  ```
---

### Slice 3 — Control Rods Change k-eff

**Goal**: Rod insertion percentage feeds into k-eff. The player can insert and
withdraw rods via console input and watch power respond in real time.

**Files touched**:
- `reactor/control_rods.go` — rod state, insert/withdraw logic, k-eff contribution
- `physics/neutronics.go` — incorporate rod worth into k-eff calculation
- `sim/engine.go` — read keyboard input concurrently alongside tick loop

**Physics introduced**:
- Each inserted rod reduces k-eff by a small worth value
- Total rod worth: fully inserted all 211 rods = approximately -0.030 delta-k
- Graphite tip flaw: inserting a rod from fully withdrawn briefly adds +0.003 to k-eff
before the absorber takes effect (modelled as a one-tick spike)
- Safety threshold: fewer than 15 rods inserted triggers a WARNING

**Controls**:
```
i   — insert 1 rod
I   — insert 10 rods
o   — withdraw 1 rod
O   — withdraw 10 rods
q   — quit
```

**Done when**:
- Inserting rods visibly reduces power over subsequent ticks
- Withdrawing rods visibly raises power
- Withdrawing below 15 rods prints a warning
- Graphite tip spike briefly appears on insertion from fully withdrawn state

**Expected output**:
```
[t=   5s]  Power:  2100.0 MW  k-eff: 1.089  Core: 310°C  Rods:  5/211  Status: WARNING
> O (withdraw 10)
WARNING: 0 rods inserted -- below safe minimum of 15
[t=   6s]  Power:  2380.0 MW  k-eff: 1.120  Core: 318°C  Rods:  0/211  Status: WARNING
> I (insert 50)
[t=   7s]  Power:  2100.0 MW  k-eff: 1.001  Core: 311°C  Rods: 50/211  Status: STABLE
```

---

### Slice 4 — The Reactor Can Die

**Goal**: When power exceeds the destruction threshold, the reactor enters MELTDOWN.
The simulation halts and prints the final moment.

**Files touched**:
- `reactor/reactor.go` — add Destroyed flag, meltdown threshold constant
- `sim/engine.go` — check meltdown condition each tick, halt on destruction

**Thresholds**:
```
Power > 3200 MW (100% nominal)  →  Status: WARNING
Power > 9600 MW (300% nominal)  →  Status: SCRAM recommended
Power > 32000 MW (1000% nominal) →  MELTDOWN — reactor destroyed
```

**Done when**:
- Leaving rods withdrawn long enough causes exponential power rise
- Power crossing 32000 MW triggers meltdown message and halts the sim
- The Chernobyl scenario (all rods withdrawn, high void) reaches meltdown
within a few seconds of simulation time

**Expected output**:
```
[t=  11s]  Power:  8400.0 MW  k-eff: 1.320  Core: 890°C  Rods:  0/211  Status: WARNING
[t=  12s]  Power: 18200.0 MW  k-eff: 1.480  Core: 1240°C  Rods:  0/211  Status: WARNING
[t=  13s]  Power: 32400.0 MW  k-eff: 1.650  Core: 1840°C  Rods:  0/211  Status: MELTDOWN

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 REACTOR DESTROYED
 1986-04-26  01:23:44
 PROMPT CRITICALITY EXCEEDED
 Peak power: 32400 MW  (1012% of nominal)
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
```

---

## File Map

```
nuclear-sim/
├── main.go                    introduced: Slice 1
├── reactor/
│   ├── reactor.go             introduced: Slice 1
│   └── control_rods.go        introduced: Slice 3
├── physics/
│   └── neutronics.go          introduced: Slice 2
└── sim/
  └── engine.go              introduced: Slice 1, expanded: Slice 2, 3, 4
```

---

## Physics Constants Used in Phase 1

| Constant | Value | Used in |
|---|---|---|
| Delayed neutron fraction (beta) | 0.0065 | Slice 2 — point kinetics |
| Prompt neutron lifetime | 0.0001 s | Slice 2 — power rate of change |
| Rod worth per rod | ~0.000142 delta-k | Slice 3 — rod k-eff contribution |
| Total rod worth (all 211) | ~0.030 delta-k | Slice 3 |
| Graphite tip k-eff spike | +0.003 delta-k | Slice 3 |
| Nominal power | 3200 MW thermal | Slice 4 — thresholds |
| Meltdown threshold | 32000 MW (1000%) | Slice 4 |

---

## Definition of Done — Phase 1

- [x] Slice 1: reactor prints static state to console
- [ ] Slice 2: power responds to k-eff each tick
- [ ] Slice 3: player can insert and withdraw rods, power responds
- [ ] Slice 4: reactor can reach meltdown and halt

When all four are complete, we proceed to Phase 2 — coolant, xenon, and
thermal feedback.

