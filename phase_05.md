# Phase 5 — Scenarios, Polish, and the Chernobyl Recreation

The goal of Phase 5 is completion. The physics are faithful, the graphics are
alive, the controls are in the player's hands. Now we tell the story — scripted
scenarios, a PWR comparison, unit tests, and the final polish that makes this
simulator worthy of the people who lived through April 26, 1986.

---

## Vertical Slices

### Slice 22 — Scenario File Loader

**Goal**: Scenarios are no longer hardcoded constructors. They are loaded from
YAML files in the scenarios/ directory. Each file defines the initial reactor
state, a sequence of scripted events, and metadata shown on the menu screen.

**Files touched**:
- `scenarios/scenario.go` — scenario struct, YAML loader
- `ui/menu.go` — read scenario files from disk, populate menu dynamically
- `main.go` — pass loaded scenario to engine

**Scenario YAML structure**:
```yaml
name: "April 26, 1986 -- The Chernobyl Accident"
description: "Deep xenon pit. 6 rods inserted. Coolant pumps slowing."
date: "1986-04-26T01:22:30"

initial_state:
power_mw: 200
k_effective: 1.001
core_temp_c: 290
fuel_temp_c: 650
coolant_flow: 7000
void_fraction: 0.08
xenon_level: 0.94
rods_inserted: 6
pressure: 65

events:
- at_sim_seconds: 30
  action: reduce_coolant_flow
  value: 500
- at_sim_seconds: 60
  action: reduce_coolant_flow
  value: 500
- at_sim_seconds: 90
  action: log
  message: "TEST BEGINNING -- turbine rundown initiated"
```

**Done when**:
- scenarios/ directory is scanned at startup
- Each YAML file appears as a menu entry with its name and description
- Scripted events fire at the correct sim time
- Adding a new YAML file adds a new scenario without changing any Go code

---

### Slice 23 — The Chernobyl Scenario

**Goal**: A fully scripted recreation of the events of April 26, 1986. The
scenario guides the simulation through the exact sequence of operator actions
and reactor responses that led to the explosion. The player watches — or
intervenes.

**File**:
- `scenarios/chernobyl.yaml`

**Scripted timeline**:
```
01:22:30  Scenario begins
        State: 200 MW, xenon pit, 6 rods inserted, pumps slowing

01:22:40  Coolant flow reduced to 6500 m3/h (pumps on turbine rundown)
        Log: "TURBINE RUNDOWN TEST INITIATED"

01:22:50  Coolant flow reduced to 5500 m3/h
        Log: "COOLANT FLOW DROPPING"

01:23:00  Void fraction rising — k-eff begins climbing
        Log: "WARNING: POWER RISING"

01:23:10  Coolant flow at 4000 m3/h
        Alarm: LOW COOLANT FLOW

01:23:20  Power climbing rapidly — 800 MW and rising
        Alarm: HIGH POWER

01:23:30  Log: "OPERATOR: AZ-5 BUTTON PRESSED"
        SCRAM initiated — all 211 rods insert

01:23:31  Graphite tip spike fires — k-eff surges
        Log: "GRAPHITE TIP SPIKE DETECTED"

01:23:32  Prompt criticality exceeded
        MELTDOWN
```

**Player interaction**:
- Player can watch passively — events fire automatically
- Player can intervene at any point: insert rods earlier, restore coolant flow
- If player prevents meltdown — alternate ending screen:
```
REACTOR STABILISED
History was changed.
In our timeline, no one intervened in time.
```
- If meltdown occurs — standard destruction screen with historical note:
```
REACTOR DESTROYED
1986-04-26  01:23:44

In the early hours of April 26, 1986, reactor No. 4 at the
Chernobyl Nuclear Power Plant was destroyed during a safety test.

31 people died in the immediate aftermath.
The exclusion zone remains today.
```

**Done when**:
- Scenario loads from chernobyl.yaml
- Timeline events fire at correct sim times
- Passive playthrough reaches meltdown naturally
- Active intervention can prevent meltdown
- Both endings display correctly

---

### Slice 24 — Safe Operation Tutorial Scenario

**Goal**: A guided scenario for a player with no nuclear knowledge. Text prompts
appear at key moments explaining what is happening and what the player should do.
By the end, the player understands how to operate the reactor safely.

**File**:
- `scenarios/tutorial.yaml`

**Tutorial sequence**:
```
Step 1  "This is your reactor. Currently stable at 1600 MW."
      "k-eff is 1.000 -- the chain reaction is self-sustaining."
      Goal: observe for 10 seconds

Step 2  "Press I to insert control rods. Watch power fall."
      Goal: reduce power below 1200 MW

Step 3  "Press O to withdraw rods. Bring power back to 1600 MW."
      Goal: restore power to 1400-1800 MW range

Step 4  "Press G to reduce coolant flow. Watch temperature rise."
      "Notice how void fraction increases -- more steam in the core."
      Goal: reduce flow to 5000 m3/h and observe

Step 5  "Press F to restore flow. Temperature will fall."
      Goal: restore flow above 7000 m3/h

Step 6  "This is the xenon gauge. At low power, xenon builds up."
      "Speed up to 10x and watch what happens to xenon over time."
      Goal: accelerate to 10x, observe xenon rising

Step 7  "You have mastered basic reactor operation."
      "Now try the Chernobyl scenario -- and see if you can stop it."
```

**Done when**:
- Tutorial prompts appear at correct moments
- Goal completion is detected and advances the tutorial
- Final screen invites player to try Chernobyl scenario

---

### Slice 25 — PWR Reactor Config

**Goal**: Add a Pressurised Water Reactor configuration as an alternative to
the RBMK. The PWR has a negative void coefficient — it is inherently
self-stabilising. Running the same Chernobyl conditions in a PWR shows the
player exactly why reactor design matters.

**Files touched**:
- `scenarios/pwr_nominal.yaml` — PWR starting state
- `scenarios/pwr_chernobyl_conditions.yaml` — same conditions as Chernobyl but in a PWR
- `physics/neutronics.go` — read void coefficient sign from reactor config

**Key differences in PWR config**:
```yaml
reactor_type: PWR
void_coefficient: -0.004    # negative -- self-stabilising
moderator: water            # water moderates AND cools
control_rod_tip: absorber   # no graphite tip flaw
nominal_power_mw: 3000
```

**Expected behaviour**:
- Running Chernobyl conditions in PWR: as coolant boils, k-eff falls,
power falls, coolant cools, void fraction drops -- self-correcting
- SCRAM in PWR: no graphite tip spike -- clean shutdown
- Player can observe side by side why RBMK was uniquely dangerous

**Scenario**:
```
"PWR -- Chernobyl Conditions"
Same initial state as April 26, 1986.
Same operator actions. Same test.

Watch what happens differently.
```

**Done when**:
- PWR scenario loads and runs with negative void coefficient
- Reducing coolant flow in PWR causes power to fall, not rise
- SCRAM in PWR shuts down cleanly without spike
- Menu shows both RBMK and PWR scenarios clearly labelled

---

### Slice 26 — Unit Tests for All Physics

**Goal**: Every physics calculation has a unit test. Tests verify that the
physics behave correctly at boundary conditions — stable equilibrium, xenon
pit, prompt criticality, graphite tip spike.

**Files touched**:
- `physics/neutronics_test.go`
- `physics/thermodynamics_test.go`
- `reactor/control_rods_test.go`
- `reactor/coolant_test.go`
- `reactor/xenon_test.go`

**Key test cases**:

```go
// neutronics_test.go
TestKEffStableAtNominal           // k-eff = 1.000 holds power steady
TestKEffAboveOneCausesPowerRise   // k-eff = 1.010 causes power to grow
TestKEffBelowOneCausesPowerFall   // k-eff = 0.990 causes power to fall
TestPromptCriticalityThreshold    // k-eff > 1.1 triggers meltdown condition
TestGraphiteTipSpike              // rod insertion from withdrawn spikes k-eff

// coolant_test.go
TestVoidFractionRisesWithTemp     // higher temp produces more steam
TestPositiveVoidCoefficient       // more void raises k-eff in RBMK
TestNegativeVoidCoefficient       // more void lowers k-eff in PWR
TestCoolantLossRaisesTemp         // zero flow causes temperature to climb

// xenon_test.go
TestXenonBuildsAtLowPower         // xenon accumulates below nominal power
TestXenonBurnsAtHighPower         // xenon consumed above nominal power
TestXenonPitSuppression           // deep xenon pit reduces k-eff significantly
TestXenonDecay                    // xenon decays with correct half-life

// control_rods_test.go
TestRodInsertionReducesKEff       // more rods = lower k-eff
TestRodWithdrawalRaisesKEff       // fewer rods = higher k-eff
TestMinimumRodWarning             // below 15 rods triggers warning
TestScramInsertsAllRods           // SCRAM sets all rods to inserting
```

**Done when**:
- `go test ./...` passes with zero failures
- All boundary conditions covered
- Physics constants documented in test comments with real-world sources

---

### Slice 27 — Final Polish

**Goal**: The simulator feels complete and respectful of its subject matter.
Visual details, performance, and presentation are refined.

**Tasks**:

**Performance**:
- Particle pool: pre-allocate particle arrays, reuse dead particles
rather than allocating new ones each tick
- Cap active particles at 2000 to maintain 60 FPS under all conditions
- Profile with `go tool pprof` and resolve any hot spots

**Visual polish**:
- Bitmap font replaced with a proper monospace font loaded from file
- Core grid cells have subtle scanline overlay for Soviet-era CRT aesthetic
- Coolant channel has a faint blue ambient glow at nominal operation
- Dashboard panel borders use consistent styling
- Meltdown debris particles linger on screen longer — slow fade to black

**Presentation**:
- Window title updates with current sim time:
`RBMK-1000 Simulator -- 1986-04-26 01:23:44`
- Screenshot key: `P` saves a PNG of the current frame to disk
- Version string displayed on menu screen

**Historical note screen**:
- Accessible from menu with `H`
- Brief text on the Chernobyl accident, its causes, and its legacy
- Names of the first responders who died

**Done when**:
- Simulator runs at stable 60 FPS under all conditions including meltdown
- All visual polish items applied
- Historical note screen accessible and complete
- Screenshot key works

---

## File Map

```
nuclear-sim/
├── main.go
├── reactor/
│   ├── reactor.go
│   ├── control_rods.go
│   ├── coolant.go
│   └── xenon.go
├── physics/
│   ├── neutronics.go
│   ├── neutronics_test.go          introduced: Slice 26
│   ├── thermodynamics.go
│   └── thermodynamics_test.go      introduced: Slice 26
├── sim/
│   ├── engine.go
│   └── events.go
├── scenarios/
│   ├── scenario.go                 introduced: Slice 22
│   ├── chernobyl.yaml              introduced: Slice 23
│   ├── tutorial.yaml               introduced: Slice 24
│   ├── pwr_nominal.yaml            introduced: Slice 25
│   └── pwr_chernobyl_conditions.yaml  introduced: Slice 25
├── particles/
│   ├── particle.go
│   ├── neutron.go
│   ├── steam.go
│   └── explosion.go
└── ui/
  ├── game.go
  ├── core_view.go
  ├── dashboard.go
  ├── menu.go
  └── assets/
      └── fonts/                  introduced: Slice 27
```

---

## Definition of Done — Phase 5

- [ ] Slice 22: scenario YAML loader works, menu populated from files
- [ ] Slice 23: Chernobyl scenario scripted, both endings implemented
- [ ] Slice 24: tutorial scenario guides new player through reactor operation
- [ ] Slice 25: PWR config demonstrates negative void coefficient
- [ ] Slice 26: all physics unit tests pass with `go test ./...`
- [ ] Slice 27: 60 FPS stable, visual polish complete, historical note screen

---

## Project Complete

When Phase 5 is done, the simulator tells the full story:

- A player with no nuclear knowledge can learn how a reactor works
- The same player can operate the Chernobyl reactor and feel the weight
of the decisions made that night
- They can try to change history — or watch it repeat
- They can compare RBMK against PWR and understand why design matters

In memory of the firefighters, operators, and liquidators of Chernobyl.
