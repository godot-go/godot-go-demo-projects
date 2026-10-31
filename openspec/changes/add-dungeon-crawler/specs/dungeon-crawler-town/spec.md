# Spec Delta

## Purpose

Defines the persistent hub between runs: the walkable town, bump-to-interact buildings, the renown economy and upgrade purchases, class unlock and selection, and the save data that is the game's only persisted state.

## ADDED Requirements

### Requirement: Town is walkable and bump opens buildings

The town SHALL be a walkable tile map on which bumping a building opens its panel and bumping the gate starts a run. The wanderer SHALL be able to walk in all eight directions. No action point system SHALL apply in town.

#### Scenario: Bump building
- **WHEN** the player moves into a building tile in town
- **THEN** that building's panel opens showing its level, cost, and effect

#### Scenario: Bump gate
- **WHEN** the player moves into the gate with a class selected
- **THEN** a run begins

#### Scenario: Diagonal walk
- **WHEN** the wanderer moves diagonally into open ground
- **THEN** the position shifts one tile on both axes

### Requirement: Renown and upgrades persist between sessions

Stored renown and purchased upgrade levels SHALL be saved to persistent storage and restored when the game starts again. This save SHALL be the only state that survives between runs.

#### Scenario: Reload
- **WHEN** the game is closed and reopened
- **THEN** renown and every building's purchased level match what they were at close

#### Scenario: Fresh install
- **WHEN** no save exists
- **THEN** renown is 0, all buildings are level 0, and exactly one starter class is unlocked

### Requirement: Upgrades are bought with renown within level caps

Each building SHALL define a maximum level and a per-level renown cost that increases with level. A purchase SHALL succeed only when the building is below its cap and stored renown covers the cost, deducting the cost and raising the level by 1.

#### Scenario: Affordable purchase
- **WHEN** the player buys an upgrade they can afford below the cap
- **THEN** renown decreases by the cost and the building level rises by one

#### Scenario: Not enough renown
- **WHEN** the player buys with insufficient renown
- **THEN** the purchase is rejected and nothing changes

#### Scenario: At cap
- **WHEN** the player buys a building already at maximum level
- **THEN** the panel reports it as maxed and no renown is spent

### Requirement: Upgrade effects apply at run start

Building effects SHALL be read when a run starts and held fixed for that run: Smithy adds Strike damage per level, Clinic grants that many starting potions, Library adds a bonus to world skill checks, and Archive extends Search range.

#### Scenario: Smithy bonus
- **WHEN** a run starts with Smithy level 2
- **THEN** the player's Strikes add +2 damage

#### Scenario: Mid-run change impossible
- **WHEN** renown or levels change during a run
- **THEN** the active run's already-applied bonuses do not change

### Requirement: Guilds unlock classes for selection

Each of the four Guild buildings SHALL gate its class: level 1 unlocks the class for selection and level 2 applies that class's skill enhancement. The roster offered at the gate SHALL contain exactly the unlocked classes, and exactly one SHALL be selected before a run starts.

#### Scenario: Locked class
- **WHEN** a class's Guild is level 0
- **THEN** the class is not selectable at the gate

#### Scenario: Unlock and pick
- **WHEN** a player buys Guild level 1 and selects that class
- **THEN** the run starts with that class's skills and stat line

### Requirement: Buildings decay until upgraded

The town shell SHALL render each building's facade as a pixel-art sprite grid (16x16 cells at 4px per cell for metered buildings, 24x16 for the gate) whose ink palette carries that building's identity roof hue. The decay ladder SHALL be tier edits on that grid, and the unmetered gate SHALL ride the whole town's wear level instead. Smoke drift and lantern pulse SHALL stay animated while the town view is displayed.

#### Scenario: Fresh-install ruins
- **WHEN** no purchases have been recorded
- **THEN** every facade reads derelict: dim palette, roof holes, boards, cracks, rubble

#### Scenario: Fully-kept cap
- **WHEN** every metered building is at its cap
- **THEN** facades read well kept: lit glass, awning, flowerboxes, lantern, chimney smoke

### Requirement: Decay tiers edit the sprite grid

A level-0 building SHALL keep a dimmed palette with roof holes punched through the gable, one window boarded, the other shattered dark, and plaster cracks; the first purchase step SHALL light a window pane; a capped building SHALL earn awning stripes under the eaves, flowerboxes, a lantern beside the door, and a stone chimney. A decayed gate SHALL pit its stone face and light guttering torches in the arch.

#### Scenario: Ruin windows
- **WHEN** a building is at level 0
- **THEN** one window is boarded and the other reads as shattered void

#### Scenario: Kept trim
- **WHEN** a building reaches its cap
- **THEN** awning stripes, flowerboxes, the lantern, and the chimney appear on the same sprite

### Requirement: Shared ground wears with upkeep

Open town ground SHALL lerp from worn lawn to lush grass with the town's whole upgrade fraction (purchased levels over the sum of caps), finished with a hash-pasted dither of grass tones and blooms, and the wander trail SHALL go from dark ruts toward a warm feathered dirt path with pebbles. Per-tile scatter details SHALL be deterministic per tile, never dancing frame to frame.

#### Scenario: Derelict ground
- **WHEN** the town is observed with no upgrades
- **THEN** open ground reads worn and the trail as cold ruts

#### Scenario: Kept ground
- **WHEN** the town is fully purchased
- **THEN** open ground reads lush with blooms and the trail as feathered warm dirt

### Requirement: Town props dress the lawn

The town shell SHALL blit pixel-art props: a stone plaza well, a market stall with crate and barrel before the Smithy, lamps flanking the gate approach, and a hash-gated tree and pine scatter across free lawn tiles. Prop placement SHALL avoid buildings (with a one-tile moat), the trail, the spawn ring, occupied prop tiles, and SHALL be deterministic. Lamp and well-water glows SHALL breathe off the shared clock.

#### Scenario: Prop placement rule
- **WHEN** props are laid out for a new town
- **THEN** no prop sits on a building, the trail, the spawn, or an occupied tile, and the layout is stable across frames

#### Scenario: Lift the greenery
- **WHEN** the town is observed fresh
- **THEN** lawns carry a visible tree and pine mix, the well stands on its stone plaza, and lamps flank the gate

### Requirement: Dungeon ground is pixel art in run view

The run view SHALL paint the dungeon ground as pixel-art ink: walls as top-lit masonry in a running bond, open cells as slabs sharing their top and left joints, and feature cells keep their flat tint. Hash scatter SHALL paste deterministic decor (moss and cracks on walls, scuff patches, rubble, or bones on floors) that never covers pattern joints. Art SHALL land as merged run rectangles, and explored-but-unseen cells SHALL dim through the painter, not the artwork.

#### Scenario: Masonry reads
- **WHEN** a lit wall is observed
- **THEN** its top edge is lighter than the face, mortar rows run full width, and brick courses show a running bond

#### Scenario: Slab joints
- **WHEN** an open floor is observed
- **THEN** its top and left edges are dark joints and the interior reads slab

#### Scenario: Stable scatter
- **WHEN** the same cell is rendered on two frames
- **THEN** its scatter and decals are identical, and decor never covers pattern joints

#### Scenario: Bounded ink
- **WHEN** a tile is rasterized into runs
- **THEN** the runs tile its 7x7 art grid exactly (3px per cell) and land within the palette

#### Scenario: Memory dimming
- **WHEN** an explored cell leaves visibility
- **THEN** the same art renders dimmed, not repainted
