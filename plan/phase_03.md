# Phase 3 — Graphics and Particles

The goal of Phase 3 is to bring the reactor to life visually. We move from the
console into a real-time graphical window powered by Ebitengine. The physics from
Phase 1 and 2 remain unchanged beneath — we are only giving them eyes.

By the end of Phase 3, the reactor breathes, glows, flows, and dies beautifully.

---

## Prerequisites

Install Ebitengine before beginning:

```bash
go get github.com/hajimehoshi/ebiten/v2
```

On macOS, Ebitengine also requires:
```bash
xcode-select --install
```

---

## Vertical Slices

### Slice 9 — Open a Window

**Goal**: Replace the console print loop with an Ebitengine window. The reactor
state is rendered as plain text inside the window. Nothing drawn yet — but the
game loop is running and the physics tick inside it.

**Files touched**:
- `ui/game.go` — implement ebiten.Game interface: Update, Draw, Layout
- `main.go` — replace engine.Run() with ebiten.RunGame()

**Ebitengine Game interface**:
```go
type Game interface {
  Update() error  // called every tick — run physics here
  Draw(*ebiten.Image) // called every frame — render here
  Layout(int, int) (int, int) // return logical screen size
}
```

**Done when**:
- A 1280x800 window opens titled "RBMK-1000 Reactor Simulator"
- Reactor state (power, k-eff, temperature, status) renders as white text
- Physics still tick — values update visibly in the window
- Window closes cleanly on escape or close button

**Expected output**:
```
[window: 1280x800, dark background]
Power:   1600.0 MW
k-eff:   1.000
Core:    285°C
Status:  STABLE
```

---

### Slice 10 — Draw the Reactor Core Grid

**Goal**: Draw the top-down cross-section of the reactor core as a grid of cells.
Each cell represents either a fuel rod channel or a control rod channel. Fuel rod
cells glow orange. Control rod cells are dark. No particles yet.

**Files touched**:
- `ui/core_view.go` — draw the core grid, colour cells by type and temperature
- `ui/game.go` — call core_view from Draw()

**Visual design**:
- Core grid: 20x20 cells centred in the left panel
- Each cell: 24x24 pixels with 2px gap
- Fuel rod cell colour: interpolate from dim orange (cool) to bright white (hot)
based on FuelTempC — range 500°C (dim) to 2000°C (white)
- Control rod cell colour: dark grey when inserted, dim orange when withdrawn
- Grid background: deep navy blue

**Colour interpolation**:
```
FuelTempC   Colour
500°C    →  #8B3A00  (dim orange)
800°C    →  #CC5500  (orange)
1200°C   →  #FF8C00  (bright orange)
1600°C   →  #FFD700  (yellow)
2000°C+  →  #FFFFFF  (white hot)
```

**Done when**:
- Core grid renders in the left panel of the window
- Fuel rod cells glow appropriately for current temperature
- Control rod cells visibly darken when inserted
- Increasing temperature (via console or direct code) causes cells to shift colour

---

### Slice 11 — Neutron Particles

**Goal**: Spawn neutron particles that fly across the core grid. Particle count,
speed, and colour reflect the current power level. At low power, a handful of
slow blue-white dots drift lazily. At high power, hundreds of fast yellow-white
streaks fill the core.

**Files touched**:
- `particles/particle.go` — base particle struct and update logic
- `particles/neutron.go` — neutron spawning, behaviour, colour by power level
- `ui/core_view.go` — draw particles over the core grid

**Particle struct**:
```go
type Particle struct {
  X, Y    float64    // position in screen space
  VX, VY  float64    // velocity pixels per tick
  Life    float64    // 1.0 = just spawned, 0.0 = dead
  Decay   float64    // how fast life drains per tick
  R, G, B, A uint8  // colour
  Size    float32
}
```

**Neutron behaviour**:
- Spawned at random fuel rod cell positions
- Travel in random directions across the core
- Bounce or die when they leave the core bounds
- Spawn rate proportional to PowerMW:
```
spawnPerTick = int(PowerMW / 100)
```
- Speed proportional to PowerMW:
```
speed = 0.5 + (PowerMW / 3200) * 3.0
```
- Colour:
```
PowerMW < 1600  →  dim blue-white
PowerMW < 3200  →  bright white
PowerMW < 9600  →  yellow-white
PowerMW > 9600  →  hot orange-white
```

**Done when**:
- Neutron particles drift across the core at nominal power
- Withdrawing rods causes particle count and speed to visibly increase
- Inserting rods reduces particle activity
- Particle activity feels calm at stable operation, frantic at high power

---

### Slice 12 — Coolant Flow and Steam Particles

**Goal**: Animate coolant flowing upward through the core. Blue particles rise
steadily at nominal flow. As void fraction increases, white steam particles mix
in. At full boil, the coolant channel is overwhelmed by steam.

**Files touched**:
- `particles/steam.go` — steam and coolant particle behaviour
- `ui/core_view.go` — draw coolant channel alongside the core grid

**Visual design**:
- Coolant channel: a vertical strip to the left of the core grid
- Blue particles rise upward at a rate proportional to CoolantFlowRate
- As VoidFraction rises, blue particles are replaced by white steam particles:
```
VoidFraction 0%    →  100% blue coolant particles
VoidFraction 50%   →  50% blue, 50% white steam
VoidFraction 100%  →  100% white steam
```
- At full steam: particles scatter chaotically rather than flowing upward cleanly

**Done when**:
- Blue coolant particles flow upward smoothly at nominal operation
- Reducing coolant flow slows the particles visibly
- Rising void fraction introduces steam particles that grow to dominate
- Full coolant loss shows chaotic white steam with no ordered flow

---

### Slice 13 — Control Rod Visualisation

**Goal**: Show control rods as dark columns descending into the core from above.
The depth of each column reflects its insertion percentage. Withdrawing a rod
causes its column to visibly retract upward.

**Files touched**:
- `ui/core_view.go` — draw rod columns overlaid on the core grid

**Visual design**:
- Each control rod channel in the grid has a dark overlay
- Overlay height = insertion percentage * cell height
- Fully inserted: entire cell dark
- Fully withdrawn: cell shows full fuel rod glow beneath
- Partial insertion: dark cap descending from top of cell

**Graphite tip flash**:
- When a rod begins inserting from fully withdrawn, briefly flash the cell
with a bright orange-white colour for 3-5 frames before the dark overlay
descends — representing the graphite tip entering the core

**Done when**:
- Inserted rods are visibly dark in the core grid
- Withdrawing rods lightens their cells gradually
- Inserting rods darkens their cells gradually
- Graphite tip flash is visible on insertion from fully withdrawn state

---

### Slice 14 — Dashboard Panels

**Goal**: Draw the right-side status panel and controls panel over the window.
Readings are colour-coded by severity. Alarms flash when thresholds are exceeded.

**Files touched**:
- `ui/dashboard.go` — draw status panel, control panel, event log strip
- `ui/game.go` — call dashboard from Draw()

**Status panel contents**:
```
REACTOR STATUS
--------------
Power     ████████░░  1600 MW
k-eff     1.000
Core      ████░░░░░░  285°C
Void      █░░░░░░░░░  3.2%
Xenon     ███░░░░░░░  12.4%
Pressure  ████░░░░░░  65 bar

Status:   STABLE
Sim Time: 1986-04-26 01:22:30
```

**Colour coding**:
```
Value in safe range    →  green text
Value approaching limit →  yellow text
Value at danger limit   →  red flashing text
```

**Alarm indicators**:
```
[ ] HIGH POWER          lights up red when Power > 3200 MW
[ ] LOW COOLANT FLOW    lights up red when Flow < 4000 m3/h
[ ] HIGH TEMPERATURE    lights up red when Core > 500°C
[ ] XENON SATURATION    lights up yellow when Xenon > 0.8
[ ] PROMPT CRITICALITY  lights up red when k-eff > 1.1
```

**Event log strip** (bottom of window):
- Last 4 events shown
- New events slide in from the right
- Colour coded: white = info, yellow = warning, red = critical

**Done when**:
- All readings display correctly and update each tick
- Colour shifts from green to yellow to red as values climb
- Alarms light up correctly at their thresholds
- Event log shows recent events scrolling in

---

### Slice 15 — The Three Visual States

**Goal**: Tie the visual atmosphere of the entire window to the reactor's status.
Stable is calm and blue. Warning bleeds red at the edges. Meltdown is chaos.

**Files touched**:
- `ui/game.go` — apply global visual state to background and vignette
- `ui/core_view.go` — adjust particle intensity by status
- `particles/explosion.go` — meltdown burst sequence

**State 1 — Stable**:
- Background: deep navy #0A0E1A
- Screen edge vignette: none
- Neutrons: slow, sparse, blue-white
- Coolant: smooth blue flow upward
- Music/sound: low hum (Phase 5)

**State 2 — Warning**:
- Background: shifts toward dark red #1A0A0A at edges
- Screen edge vignette: pulsing red glow
- Neutrons: faster, denser, shifting yellow
- Coolant: steam particles mixing in
- Dashboard: yellow and red readings flashing

**State 3 — Meltdown**:
- Frame 1: full white flash (#FFFFFF fills screen)
- Frame 2-10: explosion particles burst from core centre outward
- Hundreds of orange, yellow, white particles
- Velocity: high, outward radial direction
- Screen shake: random offset ±8px applied to all rendering
- Frame 10-60: particles fade, screen dims to black
- Frame 60+: black screen with final message centred:
```
REACTOR DESTROYED
1986-04-26  01:23:44
PROMPT CRITICALITY EXCEEDED
Peak power: 34200 MW  (1069% of nominal)

Press R to restart
```

**Explosion particle behaviour**:
```go
// particles/explosion.go
func SpawnExplosion(cx, cy float64) []Particle {
  // spawn 500 particles from core centre
  // random radial velocity: 3.0 to 12.0 pixels/tick
  // colours: white -> yellow -> orange -> red as life decays
  // lifetime: 60 ticks
  // size: 2.0 to 5.0px, shrinks with life
}
```

**Done when**:
- Stable state feels calm and industrial
- Warning state feels tense — red bleeds in, particles multiply
- Meltdown fires the full explosion sequence
- Black screen and final message display correctly
- Pressing R resets the reactor and restarts the sim

---

## File Map

```
nuclear-sim/
├── main.go                         expanded: Slice 9
├── reactor/
│   ├── reactor.go                  from Phase 1 & 2
│   ├── control_rods.go             from Phase 1 & 2
│   ├── coolant.go                  from Phase 2
│   └── xenon.go                    from Phase 2
├── physics/
│   ├── neutronics.go               from Phase 1 & 2
│   └── thermodynamics.go           from Phase 2
├── sim/
│   └── engine.go                   from Phase 1 & 2
├── particles/
│   ├── particle.go                 introduced: Slice 11
│   ├── neutron.go                  introduced: Slice 11
│   ├── steam.go                    introduced: Slice 12
│   └── explosion.go                introduced: Slice 15
└── ui/
  ├── game.go                     introduced: Slice 9
  ├── core_view.go                introduced: Slice 10, expanded: 11, 12, 13
  └── dashboard.go                introduced: Slice 14
```

---

## Tech Stack Additions in Phase 3

| Purpose | Package |
|---|---|
| Window, game loop, rendering | `github.com/hajimehoshi/ebiten/v2` |
| Drawing primitives, images | `github.com/hajimehoshi/ebiten/v2/ebitenutil` |
| Text rendering | `github.com/hajimehoshi/ebiten/v2/text` |
| Bitmap font | `golang.org/x/image/font/basicfont` |

---

## Definition of Done — Phase 3

- [ ] Slice 9: Ebitengine window opens, physics tick inside it
- [ ] Slice 10: reactor core grid drawn, fuel rods glow by temperature
- [ ] Slice 11: neutron particles fly across the core, scale with power
- [ ] Slice 12: coolant flows upward, steam appears as void fraction rises
- [ ] Slice 13: control rod columns descend into the core visibly
- [ ] Slice 14: dashboard panels display all readings with colour coding
- [ ] Slice 15: three visual states — stable, warning, meltdown explosion

When all seven are complete, the reactor is visually alive. We proceed to
Phase 4 — operator controls and interactivity. 
