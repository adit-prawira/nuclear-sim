# Phase 2 — Reactor Components

The goal of Phase 2 is to bring the reactor's feedback loops to life. The reactor
must now behave realistically — coolant heats and boils, xenon builds and decays,
temperature feeds back into power. A reactor left unattended will find its own
equilibrium, or destroy itself.

Each vertical slice builds upon the living reactor from Phase 1.

---

## Vertical Slices

### Slice 5 — Coolant Heats Up and Boils

**Goal**: Coolant temperature rises with power output. As it heats, void fraction
increases — steam bubbles form. The positive void coefficient of the RBMK means
more steam drives k-eff higher, which drives more heat, which drives more steam.
This feedback loop is the heart of Chernobyl.

**Files touched**:
- `reactor/coolant.go` — coolant temperature, void fraction, flow rate model
- `physics/thermodynamics.go` — heat balance: power in, coolant out
- `physics/neutronics.go` — incorporate void fraction into k-eff calculation

**Physics introduced**:
- Heat balance per tick: core temperature rises proportionally to power output
minus heat carried away by coolant flow
- Void fraction rises as coolant temperature approaches and exceeds boiling point
(boiling point ~285°C at nominal pressure)
- RBMK void coefficient: +0.005 delta-k per 1% increase in void fraction
- Reduced coolant flow → less heat removed → higher temperature → more voids
- Loss of coolant: if flow rate reaches 0, temperature rises unchecked

**Controls added**:
```
f   — increase coolant flow rate
F   — decrease coolant flow rate
```

**Done when**:
- Reducing coolant flow causes temperature to rise over several ticks
- Rising temperature increases void fraction visibly
- Higher void fraction pushes k-eff up, which raises power further
- Full coolant loss causes rapid temperature rise toward meltdown

**Expected output**:
```
[t=   5s]  Power: 1600 MW  k-eff: 1.000  Core: 285°C  Void:  3.2%  Flow: 8000 m3/h  Status: STABLE
> F (reduce flow)
[t=   6s]  Power: 1650 MW  k-eff: 1.003  Core: 291°C  Void:  4.1%  Flow: 6000 m3/h  Status: STABLE
[t=   7s]  Power: 1720 MW  k-eff: 1.009  Core: 299°C  Void:  5.8%  Flow: 6000 m3/h  Status: WARNING
[t=   8s]  Power: 1890 MW  k-eff: 1.021  Core: 312°C  Void:  8.2%  Flow: 6000 m3/h  Status: WARNING
```

---

### Slice 6 — Xenon Builds and Poisons the Reactor

**Goal**: Xenon-135 accumulates as a fission byproduct and suppresses the chain
reaction. At low power it builds faster than it burns — creating a xenon pit
that makes the reactor very difficult to control. This was a critical factor
in the Chernobyl accident.

**Files touched**:
- `reactor/xenon.go` — xenon concentration, production rate, decay rate
- `physics/neutronics.go` — incorporate xenon reactivity penalty into k-eff

**Physics introduced**:
- Xenon production rate proportional to power level (fission rate)
- Xenon destruction: burned off by neutron flux + natural decay (9.2 hour half-life)
- At full power: xenon burn rate exceeds production — xenon stays low
- At low power: production exceeds burn — xenon accumulates (xenon pit)
- Maximum xenon worth: approximately -2800 pcm (shuts reactor down if unchecked)
- Xenon pit recovery: operators must withdraw rods to compensate — exactly what
happened at Chernobyl

**Done when**:
- Running reactor at full power holds xenon level steady and low
- Dropping power rapidly causes xenon to climb over minutes of sim time
- Xenon level above ~0.5 forces operator to withdraw rods to maintain power
- Deep xenon pit (level > 0.9) makes it nearly impossible to hold power above 200 MW

**Expected output**:
```
[t= 120s]  Power:  200 MW  k-eff: 1.001  Xenon: 94.2%  Rods:  6/211  Status: WARNING
WARNING: Deep xenon pit -- reactor suppressed
WARNING: Only 6 rods inserted -- below safe minimum of 15
[t= 121s]  Power:  185 MW  k-eff: 0.998  Xenon: 94.8%  Rods:  6/211  Status: WARNING
```

---

### Slice 7 — Thermal Feedback Loop

**Goal**: Temperature feeds back into the physics. Hot fuel rods affect neutron
behaviour. The reactor now has a complete internal feedback system — power affects
heat, heat affects coolant, coolant affects k-eff, k-eff affects power.

**Files touched**:
- `physics/thermodynamics.go` — fuel temperature coefficient, heat transfer model
- `physics/neutronics.go` — incorporate fuel temperature into k-eff (Doppler effect)

**Physics introduced**:
- Doppler broadening: as fuel temperature rises, uranium absorbs more neutrons
without fission — a negative feedback that partially stabilises the reactor
- Fuel temperature coefficient for RBMK: approximately -0.003 pcm/°C
(partially counteracts the positive void coefficient)
- Heat transfer from fuel to coolant modelled as proportional to temperature
difference between fuel and coolant
- Fuel temperature lags behind power changes — a buffer against instant runaway

**Done when**:
- A sudden k-eff spike causes fuel temperature to rise, which partially suppresses k-eff
- The Doppler effect visibly slows power rise compared to Slice 2
- Reactor can reach a stable equilibrium at partial power without operator input
- At extreme temperatures the Doppler effect is overwhelmed by positive void coefficient

**Expected output**:
```
[t=   3s]  Power: 1640 MW  k-eff: 1.008  Core: 288°C  Fuel: 635°C  Void: 3.4%  Status: STABLE
[t=   4s]  Power: 1698 MW  k-eff: 1.005  Core: 292°C  Fuel: 648°C  Void: 3.7%  Status: STABLE
[t=   5s]  Power: 1731 MW  k-eff: 1.002  Core: 295°C  Fuel: 657°C  Void: 3.9%  Status: STABLE
[t=   6s]  Power: 1748 MW  k-eff: 1.001  Core: 296°C  Fuel: 661°C  Void: 4.0%  Status: STABLE
```
*Doppler effect slowing the rise — reactor approaching new equilibrium*

---

### Slice 8 — The Chernobyl Scenario Is Reproducible

**Goal**: Load the pre-accident reactor state from Slice 1 (ChernobylScenario
constructor) and verify that the chain of events leading to destruction unfolds
naturally from the physics — without any scripted intervention.

**Files touched**:
- `reactor/reactor.go` — ChernobylScenario() constructor (already planned in Phase 1)
- `sim/engine.go` — add scenario selection at startup

**What must happen naturally**:
1. Deep xenon pit holds power suppressed near 200 MW
2. Nearly all rods withdrawn to fight xenon
3. Coolant pumps slowing for the test raises void fraction
4. Positive void coefficient pushes k-eff above 1
5. Operator hits SCRAM — graphite tip spike fires across all rods
6. Power surges past meltdown threshold
7. Reactor destroyed

**Done when**:
- Starting with ChernobylScenario() and reducing coolant flow leads to meltdown
within ~30 seconds of sim time without any other operator input
- The sequence of events matches the historical record
- The graphite tip spike is visibly the killing blow in the event log

**Expected output**:
```
[t=   1s]  Power:  200 MW  k-eff: 1.001  Xenon: 94%  Void:  8%  Rods:  6/211  Status: WARNING
[t=   5s]  Power:  240 MW  k-eff: 1.008  Xenon: 94%  Void: 12%  Rods:  6/211  Status: WARNING
[t=   8s]  Power:  890 MW  k-eff: 1.089  Xenon: 91%  Void: 28%  Rods:  6/211  Status: WARNING
> SCRAM (AZ-5)
[t=   9s]  GRAPHITE TIP SPIKE -- k-eff surge: 1.089 -> 1.640
[t=   9s]  Power: 34200 MW  k-eff: 1.640  Core: 1840°C  Status: MELTDOWN

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 REACTOR DESTROYED
 1986-04-26  01:23:44
 PROMPT CRITICALITY EXCEEDED
 Peak power: 34200 MW  (1069% of nominal)
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
```

---

## File Map

```
nuclear-sim/
├── main.go                         from Phase 1, expanded: Slice 8
├── reactor/
│   ├── reactor.go                  from Phase 1
│   ├── control_rods.go             from Phase 1
│   ├── coolant.go                  introduced: Slice 5
│   └── xenon.go                    introduced: Slice 6
├── physics/
│   ├── neutronics.go               from Phase 1, expanded: Slice 5, 6, 7
│   └── thermodynamics.go           introduced: Slice 5, expanded: Slice 7
└── sim/
  └── engine.go                   from Phase 1, expanded: Slice 5, 8
```

---

## Physics Constants Used in Phase 2

| Constant | Value | Used in |
|---|---|---|
| RBMK void coefficient | +0.005 delta-k per 1% void | Slice 5 |
| Coolant boiling point | ~285°C at 65 bar | Slice 5 |
| Xenon-135 half-life | 9.2 hours | Slice 6 |
| Maximum xenon worth | -2800 pcm | Slice 6 |
| Doppler fuel temp coefficient | -0.003 pcm/°C | Slice 7 |
| Graphite tip k-eff spike | +0.003 delta-k per rod | Slice 8 |

---

## Definition of Done — Phase 2

- [ ] Slice 5: coolant heats, boils, void fraction rises, k-eff responds
- [ ] Slice 6: xenon builds at low power, creates xenon pit, suppresses reactor
- [ ] Slice 7: fuel temperature Doppler effect partially stabilises reactor
- [ ] Slice 8: Chernobyl scenario reaches meltdown naturally from physics alone

When all four are complete, the reactor physics are faithful enough to tell the
story of April 26, 1986. We proceed to Phase 3 — graphics and particles.
