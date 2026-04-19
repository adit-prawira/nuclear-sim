# nuclear-sim

A graphical nuclear reactor simulator written in Go.

Built to model the physics of a nuclear reactor in real-time — and to understand what went wrong at Chernobyl on April 26, 1986.

> *"Not an explosion. A nuclear excursion."*

---

## Purpose

This project exists for learning. Starting from zero knowledge of nuclear physics, the goal is to build a simulator faithful enough to:

- Understand how a nuclear chain reaction works
- Understand what makes a reactor stable or unstable
- Reproduce the conditions and operator actions that led to the Chernobyl disaster

The core physics engine is **generic** — applicable to any reactor type. The default configuration models the **RBMK-1000**, the Soviet reactor design used at
Chernobyl.

---

## Nuclear Physics Primer

### The Chain Reaction

A uranium atom struck by a neutron undergoes **fission** — it splits, releasing energy and 2–3 more neutrons. Those neutrons strike more atoms. This is the chain
reaction.

The key metric is **k-effective (k-eff)**:

| k-eff | Meaning |
|-------|---------|
| `< 1` | Reaction dying out (subcritical) |
| `= 1` | Stable, self-sustaining reaction (critical) |
| `> 1` | Reaction growing (supercritical) — danger |

### Delayed Neutrons

Not all neutrons are released instantly. A small fraction (~0.65%) arrive seconds later from decay products. This is what makes reactors *controllable* — without
delayed neutrons, reactors would be impossible to operate by hand.

### Xenon Poisoning

Fission produces Xenon-135 as a byproduct. Xenon absorbs neutrons, suppressing the reaction. At low power, xenon builds up faster than it burns off — creating a
**xenon pit** that makes it very difficult to restart or raise power. This was a key factor at Chernobyl.

### Void Coefficient

When reactor coolant heats up and begins boiling, steam bubbles (voids) form. This changes how neutrons behave. The **void coefficient** describes the effect:

- **Negative void coefficient** (most modern reactors): more steam → reaction slows → self-correcting
- **Positive void coefficient** (RBMK): more steam → reaction *speeds up* → self-reinforcing → dangerous

### Control Rods

Control rods are made of boron carbide — a material that absorbs neutrons. When inserted into the core they catch neutrons before those neutrons can cause fission, slowing or stopping the chain reaction.

| Rod Position | Effect on k-eff |
|---|---|
| Fully inserted | Maximum absorption → reactor shuts down |
| Partially inserted | Fine-tune power level |
| Fully withdrawn | No absorption → reactor runs freely |

### The Graphite Tip Flaw (RBMK)

RBMK control rods had a graphite tip at the bottom, followed by neutron-absorbing material above. When a rod is inserted from a fully withdrawn position, the
graphite tip enters the core first — **briefly increasing reactivity** before the absorber suppresses it.

At Chernobyl, all 211 rods were inserted simultaneously during SCRAM. Every graphite tip entered the core at once — causing a massive reactivity spike at the worst possible moment.

### What Happened at Chernobyl

1. Reactor running at dangerously low power for a safety test
2. Xenon buildup suppressed the reactor — operators pulled out nearly all control rods to compensate (196 of 211 withdrawn)
3. Safety test began — coolant pumps slowed, coolant started boiling
4. Positive void coefficient: more steam → k-eff rose → power surged
5. Operators hit AZ-5 (emergency SCRAM) — all 211 rods inserted simultaneously
6. Graphite tips caused a massive reactivity spike across the entire core
7. Power reached ~30,000% of nominal in seconds
8. Steam explosion — reactor destroyed

---

## Reactor Types Supported

The simulator uses a config-driven approach. Reactor-type-specific parameters live in scenario files, making it possible to model different designs and compare their behaviour.

| Reactor Type | Void Coefficient | Moderator | Status |
|---|---|---|---|
| RBMK-1000 (Chernobyl) | Positive | Graphite | Default |
| PWR (Pressurized Water) | Negative | Water | Planned |

---

## Features

- Real-time 2D graphical window (1280x800) rendered at 60 FPS via Ebitengine
- Animated reactor core — top-down cross-section with live particle effects
- Neutron particles that multiply, accelerate, and explode with power level
- Coolant visualised as animated blue flow — turns to steam as temperature rises
- Fuel rods glow orange → red → white as temperature climbs
- Control rod columns visible descending into the core
- Graphite tip flaw: brief reactivity spike on rod insertion, visible as particle burst
- Full neutronics model: k-eff, delayed neutrons, prompt criticality
- Xenon-135 buildup and decay (xenon pit)
- Thermal feedback loop
- SCRAM (AZ-5) emergency shutdown
- Three distinct visual states: Stable, Warning, Meltdown
- Meltdown sequence: white flash, particle explosion, screen shake, debris
- Scripted scenarios loaded from YAML
- Event log with sim-time timestamps
- Operator keyboard and mouse controls
- Sim speed control: pause, 1x, 5x, 10x

---

## Visual Design

### Screen Layout (1280x800)

 ```
  +-------------------------------+---------------------+-----------------------+
  |                               |                     |                       |
  |                               |   REACTOR STATUS    |   CONTROL RODS        |
  |                               |                     |                       |
  |       REACTOR CORE            |  Power   ########.  |   Rod 1  ####....     |
  |    (top-down cross-section)   |  Temp    ######.... |   Rod 2  ####....     |
  |                               |  Void    ##........ |   ...                 |
  |  - neutron particles          |  Xenon   ###....... |   [INSERT ALL]        |
  |  - fuel rod glow              |  k-eff:  1.002      |   [WITHDRAW ALL]      |
  |  - control rod columns        |  Status: STABLE     |   [SCRAM  AZ-5]       |
  |  - coolant flow animation     |                     |                       |
  |  - steam particles            |  Sim Time:          |   COOLANT FLOW        |
  |  - explosion on meltdown      |  1986-04-26         |   ^ ^ ^ animated      |
  |                               |  01:22:30           |   Flow: 8000 m3/h     |
  +-------------------------------+---------------------+-----------------------+
  |  EVENT LOG                                                                  |
  |  01:22:28  Coolant flow reduced to 8000 m3/h                                |
  |  01:22:15  196 control rods withdrawn -- WARNING: below safe minimum        |
  +------------------------------------------------------------------------------+
  ```

### Colour Language

| Colour | Meaning |
|---|---|
| Deep blue | Cold, safe, nominal coolant |
| Cyan flowing upward | Active coolant flow |
| Orange | Fuel rods at normal temperature |
| Yellow | Elevated temperature / rising power |
| Red | Danger — high heat or alarm state |
| White burst | Neutron flash / criticality event |
| Dark grey | Control rod inserted |
| Green text | Normal status readings |
| Yellow text | Warning readings |
| Red flashing text | Alarm / critical state |

### Three Visual States

| State | Trigger | Visual |
|---|---|---|
| Stable | k-eff near 1.0 | Dark blue, slow neutrons, soft orange glow |
| Warning | k-eff > 1.05 or temp rising | Red edges, faster neutrons, steam particles appear |
| Meltdown | Prompt criticality | White flash, explosion, screen shake, debris, blackout |

---

## Architecture

```
nuclear-sim/
├── main.go
├── reactor/
│   ├── reactor.go        # core state: power, temperature, pressure
│   ├── core.go           # fuel rods, neutron flux, k-eff calculation
│   ├── control_rods.go   # rod positions, graphite tip effect
│   ├── coolant.go        # water flow, void fraction, boiling
│   └── xenon.go          # xenon poisoning buildup/decay
├── physics/
│   ├── neutronics.go     # k-eff, delayed neutrons, prompt criticality
│   └── thermodynamics.go # heat transfer, steam generation
├── sim/
│   ├── engine.go         # time-step simulation loop
│   └── events.go         # event log, alarms, triggers
├── scenarios/
│   └── chernobyl.yaml    # April 26, 1986 scenario
├── particles/
│   ├── particle.go       # base particle: position, velocity, lifetime, colour
│   ├── neutron.go        # neutron particle behaviour
│   ├── steam.go          # steam/coolant particle behaviour
│   └── explosion.go      # meltdown particle burst
└── ui/
  ├── game.go           # ebiten Game interface — main render loop
  ├── core_view.go      # draw reactor core grid + particles
  ├── dashboard.go      # draw status panels and event log
  └── assets/
      └── fonts/        # bitmap font for Soviet-era terminal aesthetic

```
---

## Tech Stack

| Purpose | Choice |
|---|---|
| Language | Go 1.22+ |
| Graphics & Window | `github.com/hajimehoshi/ebiten/v2` |
| Text rendering | `golang.org/x/image/font` + bitmap font |
| Scenarios / Config | `gopkg.in/yaml.v3` |
| Testing | standard `testing` package |

---

## Milestones

### Phase 1 — The Core Engine
- [x] Slice 1: reactor prints static state to console
- [x] Slice 2: power responds to k-eff each tick
- [x] Slice 3: player can insert and withdraw rods, power responds
- [x] Slice 4: reactor can reach meltdown and halt

### Phase 2 — Reactor Components
- [x] Slice 5: coolant heats, boils, void fraction rises, k-eff responds
- [x] Slice 6: xenon builds at low power, creates xenon pit, suppresses reactor
- [ ] Slice 7: fuel temperature Doppler effect partially stabilises reactor
- [ ] Slice 8: Chernobyl scenario reaches meltdown naturally from physics alone

### Phase 3 — Graphics and Particles
- [ ] Slice 9: Ebitengine window opens, physics tick inside it
- [ ] Slice 10: reactor core grid drawn, fuel rods glow by temperature
- [ ] Slice 11: neutron particles fly across the core, scale with power
- [ ] Slice 12: coolant flows upward, steam appears as void fraction rises
- [ ] Slice 13: control rod columns descend into the core visibly
- [ ] Slice 14: dashboard panels display all readings with colour coding
- [ ] Slice 15: three visual states — stable, warning, meltdown explosion

### Phase 4 — Operator Controls and Interactivity
- [ ] Slice 16: keyboard controls rod insertion and withdrawal
- [ ] Slice 17: keyboard controls coolant flow rate
- [ ] Slice 18: sim speed and pause controls work
- [ ] Slice 19: SCRAM triggers correctly — kills reactor in Chernobyl scenario
- [ ] Slice 20: all alarms trigger at correct thresholds
- [ ] Slice 21: scenario selection menu shown at startup

### Phase 5 — Scenarios, Polish, and the Chernobyl Recreation
- [ ] Slice 22: scenario YAML loader works, menu populated from files
- [ ] Slice 23: Chernobyl scenario scripted, both endings implemented
- [ ] Slice 24: tutorial scenario guides new player through reactor operation
- [ ] Slice 25: PWR config demonstrates negative void coefficient
- [ ] Slice 26: all physics unit tests pass with `go test ./...`
- [ ] Slice 27: 60 FPS stable, visual polish complete, historical note screen

---

## Key Physics Constants

| Constant | Value | Meaning |
|---|---|---|
| Void coefficient (RBMK) | `+4.7 to +8 pcm/% void` | k-eff rise per % void fraction |
| Delayed neutron fraction (β) | `0.0065` | Fraction of neutrons that are delayed |
| Xenon-135 half-life | `~9.2 hours` | Decay rate of xenon poisoning |
| Prompt neutron lifetime | `~0.0001 s` | How fast prompt neutrons travel |
| Nominal RBMK power | `3200 MW thermal` | Full power baseline |
| RBMK control rods | `211 total` | Minimum safe insertion: 15 rods |
| Graphite tip length | `~10% of rod length` | Extent of positive reactivity spike on insertion |

---

## Disclaimer

This is an educational simulator. The physics models are simplified for learning purposes and are not suitable for any real engineering application.

