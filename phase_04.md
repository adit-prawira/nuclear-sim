# Phase 4 — Operator Controls and Interactivity

The goal of Phase 4 is to put the reactor in the player's hands. The simulation
becomes a game — the player operates the reactor, responds to alarms, and decides
when to SCRAM. Every control has a visible consequence in the physics and the
graphics.

By the end of Phase 4, a player can sit down, run the reactor, make mistakes,
and watch Chernobyl happen.

---

## Vertical Slices

### Slice 16 — Keyboard Controls for Control Rods

**Goal**: The player controls rod insertion and withdrawal via keyboard inside
the Ebitengine window. Rod movements are reflected immediately in the core grid
visual and in the physics.

**Files touched**:
- `ui/game.go` — read keyboard input in Update()
- `reactor/control_rods.go` — expose insert/withdraw methods

**Key bindings**:
```
I        — insert 1 rod group (10 rods)
O        — withdraw 1 rod group (10 rods)
Shift+I  — insert all rods
Shift+O  — withdraw all rods
```

**Rules**:
- Rod movement is not instant — rods move at a fixed rate of 0.4 m/s
(realistic RBMK insertion speed)
- Each keypress queues a rod movement — held key continues movement
- Visual: rod columns in the core grid animate downward/upward smoothly
- Warning displayed on dashboard when rods fall below 15 inserted

**Done when**:
- Pressing I/O moves rods and core grid responds visually
- Physics responds — k-eff and power change as rods move
- Rod count on dashboard updates in real time
- Warning appears below 15 rods

---

### Slice 17 — Coolant Flow Control

**Goal**: The player adjusts coolant flow rate via keyboard. Changing flow
affects heat removal, coolant temperature, void fraction, and ultimately k-eff.
The coolant panel animates to reflect the change.

**Files touched**:
- `ui/game.go` — read coolant flow keys in Update()
- `reactor/coolant.go` — expose flow rate adjustment methods

**Key bindings**:
```
F        — increase coolant flow by 500 m3/h
G        — decrease coolant flow by 500 m3/h
Shift+F  — restore nominal flow (8000 m3/h)
```

**Rules**:
- Flow rate range: 0 to 14000 m3/h
- Below 2000 m3/h: WARNING — LOW COOLANT FLOW alarm triggers
- At 0 m3/h: coolant loss — temperature climbs rapidly
- Coolant particle animation speed scales with flow rate visibly

**Done when**:
- Reducing flow raises coolant temperature over subsequent ticks
- Steam particles increase visibly as void fraction climbs
- Alarm triggers at low flow threshold
- Restoring flow gradually brings temperature back down

---

### Slice 18 — Sim Speed Control

**Goal**: The player can pause, slow, and accelerate simulation time. Useful
for watching slow processes like xenon buildup, or for slowing down critical
moments.

**Files touched**:
- `sim/engine.go` — expose speed multiplier
- `ui/game.go` — read speed control keys, display current speed on dashboard

**Key bindings**:
```
Space      — pause / resume
1          — 1x speed (1 sim-second per real second)
2          — 5x speed
3          — 10x speed
4          — 30x speed (for xenon buildup observation)
```

**Display**:
```
Sim Speed:  [PAUSED]
Sim Speed:  1x
Sim Speed:  5x  >>
Sim Speed:  10x >>>
Sim Speed:  30x >>>>
```

**Done when**:
- Space pauses and resumes the simulation cleanly
- Speed keys change the tick rate visibly
- At 30x speed, xenon buildup is observable in under a minute of real time
- Dashboard shows current speed at all times

---

### Slice 19 — SCRAM Button (AZ-5)

**Goal**: The player can trigger an emergency SCRAM, inserting all 211 rods
simultaneously. In the Chernobyl scenario, this is the killing blow — the
graphite tip spike fires across all rods at once.

**Files touched**:
- `ui/game.go` — read SCRAM key, confirm dialog
- `reactor/control_rods.go` — SCRAM logic, graphite tip spike across all rods
- `ui/dashboard.go` — SCRAM button visual, confirmation prompt

**Key binding**:
```
Enter    — trigger SCRAM (AZ-5)
```

**SCRAM sequence**:
1. Player presses Enter
2. Dashboard shows confirmation: `INITIATE SCRAM? [Enter] confirm  [Esc] cancel`
3. Player confirms — all 211 rods begin inserting simultaneously
4. Each rod fires graphite tip spike as it enters: +0.003 delta-k per rod
5. Total spike: up to +0.633 delta-k if all rods fully withdrawn
6. Spike lasts 2 ticks then absorber takes over — k-eff begins falling
7. If reactor was already near criticality — meltdown fires before absorber wins

**Visual during SCRAM**:
- All rod columns in core grid animate downward at once
- Brief orange flash across all rod cells (graphite tip)
- If spike causes meltdown — explosion sequence fires (Slice 15)
- If SCRAM succeeds — reactor power falls, status moves to SCRAM then STABLE

**Done when**:
- SCRAM at nominal operation successfully shuts down the reactor
- SCRAM in Chernobyl scenario triggers meltdown via graphite tip spike
- Confirmation dialog prevents accidental SCRAM
- Visual sequence is dramatic and clear

---

### Slice 20 — Alarm System

**Goal**: Alarms trigger automatically when reactor parameters cross defined
thresholds. Alarms are visible on the dashboard, logged in the event log, and
reflected in the visual atmosphere of the window.

**Files touched**:
- `sim/events.go` — alarm definitions, threshold checks, event emission
- `ui/dashboard.go` — alarm panel, flashing indicators
- `ui/game.go` — tie alarm state to visual atmosphere

**Alarm definitions**:

| Alarm | Trigger Condition | Severity |
|---|---|---|
| HIGH POWER | Power > 3200 MW | WARNING |
| CRITICAL POWER | Power > 9600 MW | CRITICAL |
| LOW COOLANT FLOW | Flow < 4000 m3/h | WARNING |
| COOLANT LOSS | Flow = 0 | CRITICAL |
| HIGH CORE TEMP | Core > 500°C | WARNING |
| CRITICAL TEMP | Core > 1200°C | CRITICAL |
| XENON SATURATION | Xenon > 80% | WARNING |
| LOW ROD INSERTION | Rods < 15 | WARNING |
| PROMPT CRITICALITY | k-eff > 1.1 | CRITICAL |

**Visual behaviour**:
- WARNING alarms: amber indicator, yellow event log entry
- CRITICAL alarms: red flashing indicator, red event log entry
- Multiple CRITICAL alarms active: screen edge vignette intensifies
- All alarms cleared: indicators return to dim grey

**Done when**:
- All alarms trigger at correct thresholds
- Alarms appear in event log with sim timestamp
- Dashboard alarm panel reflects current alarm state
- Visual atmosphere responds to alarm severity

---

### Slice 21 — Scenario Selection Screen

**Goal**: Before the simulation begins, the player selects a starting scenario
from a simple menu screen. Two scenarios available in Phase 4: nominal operation
and the Chernobyl pre-accident state.

**Files touched**:
- `ui/menu.go` — scenario selection screen
- `ui/game.go` — show menu before starting sim, handle selection

**Menu design**:
```
+--------------------------------------------------+
|                                                  |
|     RBMK-1000 NUCLEAR REACTOR SIMULATOR          |
|     Chernobyl Nuclear Power Plant -- Unit 4      |
|                                                  |
|     Select Scenario:                             |
|                                                  |
|     [1]  Normal Operation                        |
|          Reactor at stable 1600 MW               |
|          Learn to operate the reactor            |
|                                                  |
|     [2]  April 26, 1986 -- 01:22:30              |
|          Deep xenon pit. 6 rods inserted.        |
|          Coolant pumps slowing. Test beginning.  |
|          Recreate the Chernobyl accident.        |
|                                                  |
|     [Esc] Quit                                   |
|                                                  |
+--------------------------------------------------+
```

**Done when**:
- Menu renders before simulation starts
- Pressing 1 starts nominal operation
- Pressing 2 starts Chernobyl scenario
- Selected scenario initialises correct reactor state

---

## File Map

```
nuclear-sim/
├── main.go                         from Phase 1-3
├── reactor/
│   ├── reactor.go                  from Phase 1-3
│   ├── control_rods.go             expanded: Slice 16, 19
│   ├── coolant.go                  expanded: Slice 17
│   └── xenon.go                    from Phase 2
├── physics/
│   ├── neutronics.go               from Phase 1-3
│   └── thermodynamics.go           from Phase 2-3
├── sim/
│   ├── engine.go                   expanded: Slice 18
│   └── events.go                   expanded: Slice 20
├── particles/
│   ├── particle.go                 from Phase 3
│   ├── neutron.go                  from Phase 3
│   ├── steam.go                    from Phase 3
│   └── explosion.go                from Phase 3
└── ui/
  ├── game.go                     expanded: Slice 16, 17, 18, 19, 20, 21
  ├── core_view.go                from Phase 3
  ├── dashboard.go                expanded: Slice 19, 20
  └── menu.go                     introduced: Slice 21
```

---

## Key Bindings Summary

| Key | Action |
|---|---|
| I / O | Insert / withdraw rod group |
| Shift+I / Shift+O | Insert all / withdraw all rods |
| F / G | Increase / decrease coolant flow |
| Shift+F | Restore nominal coolant flow |
| Space | Pause / resume |
| 1 / 2 / 3 / 4 | Sim speed 1x / 5x / 10x / 30x |
| Enter | SCRAM (AZ-5) — with confirmation |
| Esc | Cancel confirmation / return to menu |
| R | Restart after meltdown |

---

## Definition of Done — Phase 4

- [ ] Slice 16: keyboard controls rod insertion and withdrawal
- [ ] Slice 17: keyboard controls coolant flow rate
- [ ] Slice 18: sim speed and pause controls work
- [ ] Slice 19: SCRAM triggers correctly — kills reactor in Chernobyl scenario
- [ ] Slice 20: all alarms trigger at correct thresholds
- [ ] Slice 21: scenario selection menu shown at startup

When all six are complete, the simulator is fully playable. We proceed to
Phase 5 — scenarios, polish, and the Chernobyl recreation.
