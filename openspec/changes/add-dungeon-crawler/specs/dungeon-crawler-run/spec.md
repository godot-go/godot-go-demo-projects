# Spec Delta

## Purpose

Defines one attempt of the dungeon crawler: procedural floor generation, fog of war, interactable dungeon features, monster rosters and tiering, the win and loss conditions across floors 1 to 5, and the renown payout that ends the run.

## ADDED Requirements

### Requirement: Runs are generated fresh and discarded on completion

Starting a run SHALL procedurally generate floor 1 from a seed, and the dungeon state SHALL NOT be persisted between runs. Ending a run in death or victory SHALL discard the dungeon entirely.

#### Scenario: Fresh floor
- **WHEN** a run starts
- **THEN** floor 1 is generated and the player begins on an entrance tile with no monsters adjacent

### Requirement: Traps stay hidden until spotted

Armed traps SHALL NOT be visible until a Search succeeds at spotting one. The Search verb's Perception roll SHALL also probe for armed traps within its reach against a spot DC, revealing the nearest hidden thing — door or trap — on a success. A spotted trap SHALL stay rendered (dimmed once out of sight) until disarmed or triggered; unspotted traps SHALL still trigger when walked into.

#### Scenario: Hidden trap no-ops the eye
- **WHEN** a run generates an armed trap the player has not searched near
- **THEN** the trap renders as plain floor

#### Scenario: Search spots the trap
- **WHEN** Search succeeds with an armed trap inside reach
- **THEN** the trap becomes spotted, logs the roll against the spot DC, and renders from then on (dimmed when out of sight)

#### Scenario: Unspotted still bites
- **WHEN** the player walks onto an unspotted armed trap
- **THEN** it triggers exactly as before

### Requirement: Visible hostile creatures are outlined

Every living enemy inside the player's current sight SHALL be rendered with a hostile outline; creatures remembered from fog-of-war SHALL NOT carry one.

#### Scenario: Outline while seen
- **WHEN** a living enemy stands in a visible tile
- **THEN** its tile shows the outline for as long as it stays visible

#### Scenario: Memory keeps no outline
- **WHEN** the same tile falls back out of sight
- **THEN** no enemy outline renders there

#### Scenario: No carryover
- **WHEN** a new run starts after a previous run ended
- **THEN** the layout is generated anew and none of the previous run's dungeon state applies

### Requirement: Floors are connected room layouts

Each floor SHALL be a bounded grid of rooms connected by corridors such that every room and the stairs down are reachable from the entrance. Floor count per run SHALL be 5.

#### Scenario: Reachability
- **WHEN** a floor is generated
- **THEN** a walkable path exists from the entrance tile to the stairs and to every room

### Requirement: Visibility reveals around the player and remembers exploration

Tiles beyond the player's line-of-sight radius SHALL be hidden. Once seen, a tile SHALL remain visible as explored terrain for the rest of the run. Entities SHALL only be displayed while visible.

#### Scenario: Unseen hidden
- **WHEN** a tile has never been within the player's sight
- **THEN** it is rendered as unknown

#### Scenario: Explored memory
- **WHEN** a previously seen tile leaves sight range
- **THEN** it renders as explored terrain without live entities

### Requirement: Stairs descend one floor

Bumping the stairs down SHALL generate and enter the next floor. On floor 5 there SHALL be no usable stairs; the boss guards the exit instead.

#### Scenario: Descend
- **WHEN** the player bumps stairs on floor N where N is less than 5
- **THEN** floor N+1 is generated and the player enters it at its entrance

### Requirement: Locked doors seal an optional vault

Each floor SHALL attempt to seal one away room behind a locked door: never the entrance or stairs room, never a route the floor needs. The seal SHALL keep exactly one doorway on the room's border — never inside the room — with surplus doorways walled only when provably harmless.

#### Scenario: The lock never fences the route
- **WHEN** a floor generates
- **THEN** every tile reachable from the entrance stays reachable without opening the locked door, and the sealed room's interior touches only its locked doorway

#### Scenario: Loot waits inside
- **WHEN** an optional room is sealed
- **THEN** the floor's potion rests inside it and the weapon joins it on roughly half the floors

#### Scenario: Seal skipped rather than unsafe
- **WHEN** no room can be safely sealed on a floor
- **THEN** that floor has no locked door

### Requirement: Trained world skills open locked doors

Bumping a locked door SHALL spend 1 action point and attempt the bumper's trained world skill — Athletics forces it, Thievery picks the lock; classes trained in neither cannot open it. Failure keeps the door locked, success opens it permanently.

#### Scenario: Force open
- **WHEN** a Fighter bumps a locked door and the Athletics check meets the DC
- **THEN** the door becomes open and passable

#### Scenario: Pick open
- **WHEN** a Rogue bumps a locked door and the Thievery check meets the DC
- **THEN** the door becomes open and the log reads as lockpicking

#### Scenario: Locked out of the treasure
- **WHEN** a class trained in neither skill bumps a locked door
- **THEN** the door stays shut and no path exists through it

#### Scenario: Failed forcing
- **WHEN** the check fails
- **THEN** the door stays locked and the action point is consumed

### Requirement: Traps damage on crossing until disabled

An armed trap on a floor tile SHALL damage any creature entering it once. A Thievery check (trained-only) on bumping or targeting the trap SHALL disable it permanently before it can trigger.

#### Scenario: Triggered
- **WHEN** a creature enters an armed trap's tile
- **THEN** it takes the trap's damage and the trap is spent

#### Scenario: Disabled
- **WHEN** a Thievery check on an armed trap succeeds
- **THEN** the trap is disarmed and crossing it is safe

### Requirement: Hidden doors are found by searching

Hidden doors SHALL occupy the map as plain wall until a Search action (1 point, Perception check) within range reveals them; the Archive upgrade SHALL widen the Search radius. Revealed doors behave as openable doors.

#### Scenario: Search reveals
- **WHEN** the player Searches and a hidden door is within Search range and the check succeeds
- **THEN** the door becomes visible and usable

#### Scenario: Nothing found
- **WHEN** Search resolves with no hidden door in range
- **THEN** the action is consumed and the log reports nothing found

### Requirement: Potions restore hit points

Potions SHALL appear as floor pickups; stepping on one adds it to the carried count, and quaffing one SHALL restore a fixed amount of hit points without exceeding the maximum.

#### Scenario: Pickup
- **WHEN** the player moves onto a potion
- **THEN** the carried potion count increases and the tile is left empty

#### Scenario: Quaff
- **WHEN** the player quaffs a potion while damaged
- **THEN** hit points are restored up to the maximum and one potion is consumed

### Requirement: Monsters match the floor tier

Floor N SHALL spawn monsters drawn from the bestiary scaled to that floor's tier, with tougher species and elite scaling on deeper floors. The floor-5 boss SHALL be a single elite monster with distinct stats.

#### Scenario: Tiered spawn
- **WHEN** floor 1 is generated
- **THEN** only weakest-tier species appear, and floor 4 layouts include only deeper-tier species

### Requirement: Death loses and boss kills win the run

A run SHALL be lost when the player's hit points reach 0 and won when the floor-5 boss dies. In either case the run SHALL present its end screen and stop accepting actions.

#### Scenario: Death
- **WHEN** the player's hit points reach 0
- **THEN** the run enters the lost state and no further actions resolve

#### Scenario: Boss slain
- **WHEN** the floor-5 boss dies
- **THEN** the run enters the won state

### Requirement: Runs pay renown on completion

At run end, win or loss, renown SHALL be awarded from kills by monster level, floors descended, and first-time boss kills, added to the player's stored renown total.

#### Scenario: Death payout
- **WHEN** a run ends in death
- **THEN** renown is computed from that run's deeds and added to the stored total

#### Scenario: Victory bonus
- **WHEN** a run ends by killing the boss for the first time
- **THEN** the payout includes the first-kill bonus

### Requirement: Spell slots refresh when descending

Each floor stands in for a day of PF2e preparation: when the party descends, every Wizard spell slot SHALL refill to its level's total, with the log noting the refit. Non-wizards are unaffected.

#### Scenario: Descending refills the book
- **WHEN** a Wizard who has spent spell slots descends to the next floor
- **THEN** all of that Wizard's slots read full again

### Requirement: The spellbook modal opens on the B key

A dedicated hotkey SHALL open a pausing, bordered spellbook overlay that lists the Wizard's known spells with their costs, marks the prepared entry, and summarizes remaining slots. It pauses the game like the other modals, closes on the same hotkey, Escape, or Enter, and is advertised in the help overlay and the hotkey footer.

#### Scenario: Wizard readies a spell mid-run
- **WHEN** the spellbook is open during a run and a number key matching a known spell is pressed
- **THEN** that spell becomes the prepared entry for the class skill verb at no cost and the list reflects the change

#### Scenario: Town preview
- **WHEN** the spellbook is opened in town
- **THEN** it shows the level 1 lineup the selected class would begin the run with

#### Scenario: Non-wizard peek
- **WHEN** a non-Wizard opens the spellbook
- **THEN** the overlay states that only a Wizard keeps a book and no preparation keys apply

#### Scenario: Pause semantics
- **WHEN** the spellbook is open
- **THEN** no other verb advances the game until it is dismissed

### Requirement: Knock swings locked doors open

A level-2 Wizard SHALL learn Knock: a free, one-action cantrip that permanently opens one locked door within a step with no roll and no spell slot. With no locked door in reach the cast SHALL fizzle and commit neither the action nor a slot.

#### Scenario: Magical unlock
- **WHEN** the prepared spell is Knock and a locked door stands adjacent
- **THEN** the door swings open permanently for 1 action with no slot spent

#### Scenario: No door in reach
- **WHEN** the prepared spell is Knock and no locked door is within a step
- **THEN** nothing opens and neither the action nor a slot is spent

### Requirement: Spells destroy armed traps

Every damaging spell's reach also touches armed traps. Aimed spells (Ray of Frost, Electric Arc, Magic Missile) SHALL destroy the nearest armed trap under their foe-reach geometry, but only when no foe stands in reach; area blasts (Burning Hands, Fire Burst) SHALL destroy traps caught in their radius alongside foes. One hit destroys with no roll; a leveled spell spends its slot. With neither in reach the cast SHALL fizzle and commit nothing.

#### Scenario: Aimed spell falls back to a trap
- **WHEN** the Wizard casts an aimed spell with no foe in reach and an armed trap under its reach
- **THEN** the trap is destroyed for the spell's actions (slot included for spells with slots)

#### Scenario: Blast catches traps with foes
- **WHEN** the Wizard casts a fire blast with foes and armed traps in its radius
- **THEN** the foes take damage and the traps are destroyed in the same cast

#### Scenario: Foe outranks the trap
- **WHEN** the Wizard casts an aimed spell with both a foe and an armed trap in reach
- **THEN** the spell strikes the foe and the trap stays armed

#### Scenario: Nothing in reach
- **WHEN** the Wizard casts a spell with neither foe nor armed trap in reach
- **THEN** no action, no slot, and no feature changes

### Requirement: Figures animate on the dungeon view

Visible characters and dungeon features with a figure SHALL render as animated figures with an idle motion of their own rather than flat tiles, and the floor-5 boss SHALL draw larger than ordinary monsters. Idle motion SHALL desynchronize between entities. Explored-but-unseen feature tiles SHALL render as dimmed, frozen memories. Town SHALL draw the selected class's avatar in place of the player marker.

#### Scenario: Idle while visible
- **WHEN** the run view repaints across consecutive frames with no player input
- **THEN** every visible figure keeps animating its idle

#### Scenario: Frozen dimmed memory
- **WHEN** a tile carrying a feature leaves sight range after having been seen
- **THEN** its figure stays drawn but dimmed and motionless

#### Scenario: Desynced mobs
- **WHEN** two monsters of the same species are visible together
- **THEN** their idle cycles are offset from each other

#### Scenario: Town avatar
- **WHEN** the town view is displayed
- **THEN** the selected class's animated avatar stands on the player's tile
