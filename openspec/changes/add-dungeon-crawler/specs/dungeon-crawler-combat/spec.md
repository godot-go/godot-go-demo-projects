# Spec Delta

## Purpose

Defines the Pathfinder-2e-flavored resolution rules of the dungeon crawler: the shared three-action turn economy, bump and ranged attacks, class skills, d20 proficiency checks, and experience-based leveling.

## ADDED Requirements

### Requirement: Player turns grant three actions

Each player turn SHALL grant exactly 3 action points, spendable on Stride, Strike, or a class skill. Unspent actions SHALL NOT carry to a later turn, and the turn SHALL end automatically when points reach 0 or when the player ends it early.

#### Scenario: Spending exhausts the turn
- **WHEN** the player spends 3 action points in any combination
- **THEN** the player's turn ends and enemies act

#### Scenario: Early end
- **WHEN** the player ends the turn with points remaining
- **THEN** the remaining points are forfeited and enemies act

#### Scenario: No actions remain
- **WHEN** an action is requested while 0 points remain
- **THEN** the request is rejected with a message and state is unchanged

### Requirement: Stride moves one tile per action

Stride SHALL move the player exactly one tile in the chosen direction — orthogonal or diagonal — when the destination tile is walkable. The player MUST NOT enter a tile occupied by a wall or an enemy. A diagonal stride MUST NOT clip through wall corners: it is blocked when both orthogonal neighbours on the corner are unwalkable.

#### Scenario: Walkable destination
- **WHEN** the player Strides toward an empty floor tile
- **THEN** the player occupies that tile and 1 action point is spent

#### Scenario: Blocked destination
- **WHEN** the player Strides toward a wall or an enemy-occupied tile
- **THEN** the player does not move and no action point is spent

#### Scenario: Clipped corner
- **WHEN** the player Strides diagonally while both orthogonal neighbours of that corner are unwalkable
- **THEN** the player does not move and no action point is spent

#### Scenario: Diagonal bump strikes
- **WHEN** the player moves diagonally into a tile occupied by an enemy
- **THEN** the move resolves as a Strike against that enemy like any orthogonal bump

### Requirement: Bump attack resolves as a Strike

Moving into a tile occupied by an enemy SHALL spend 1 action point and resolve a Strike against that enemy without moving the player. The roll SHALL be d20 + the player's proficiency bonus + ability modifier against the enemy's Armor Class, and the result SHALL be reported to the combat log.

#### Scenario: Hit
- **WHEN** the Strike roll total meets or exceeds the enemy's AC
- **THEN** the enemy takes the weapon's damage dice and the log shows the roll and damage

#### Scenario: Miss
- **WHEN** the roll total is below the enemy's AC
- **THEN** the enemy takes no damage and the log shows the failed roll

#### Scenario: Critical and fumble
- **WHEN** the d20 shows a natural 20
- **THEN** the Strike hits and its damage dice are doubled
- **AND WHEN** the d20 shows a natural 1
- **THEN** the Strike misses regardless of bonuses

### Requirement: Multiple attacks take escalating penalties

Strikes made on the same turn SHALL take a −5 penalty for the second Strike and a −10 penalty for the third and any later Strike. The penalty SHALL reset at the start of each of the player's turns.

#### Scenario: Second Strike is penalized
- **WHEN** the player makes a second Strike in one turn
- **THEN** its attack roll uses a −5 penalty

#### Scenario: Third Strike is penalized further
- **WHEN** the player makes a third Strike in one turn
- **THEN** its attack roll uses a −10 penalty

#### Scenario: Penalty resets
- **WHEN** a new player turn begins
- **THEN** the next Strike rolls with no multiple attack penalty

### Requirement: Class skills spend action points

Each class SHALL expose its defined skills, each costing a fixed action cost of 1 or 2 points, and a skill MAY only be used if the class provides it and enough points remain. Skills SHALL be resolved immediately with their stated effect reported to the combat log.

#### Scenario: Power Attack
- **WHEN** a Fighter with at least 2 points remaining uses Power Attack
- **THEN** a single Strike resolves at 2 points with the Fighter's enhanced damage

#### Scenario: Lay on Hands
- **WHEN** a Cleric uses Lay on Hands
- **THEN** the Cleric heals its defined amount up to its maximum and 1 point is spent

#### Scenario: Skill unavailable
- **WHEN** a skill the player's class does not provide is requested
- **THEN** the request is rejected and state is unchanged

### Requirement: Checks use d20 with proficiency against a DC

Skill checks against difficulty classes (locked doors, traps, hidden doors) SHALL roll d20 + proficiency bonus + ability modifier against the feature's DC, where a trained skill contributes level + 2 and an untrained skill contributes the ability modifier only. A check whose skill is restricted to trained SHALL fail automatically when the player is untrained.

#### Scenario: Trained success
- **WHEN** a trained player's check total meets the DC
- **THEN** the check succeeds and its effect applies

#### Scenario: Untrained attempt on trained-only skill
- **WHEN** a player untrained in the required skill attempts a trained-only check
- **THEN** the attempt fails without granting any retry cost refund and the log states why

### Requirement: Ranged attacks strike the nearest enemy in reach

A ranged Strike SHALL target the nearest enemy within its reach in tiles, measured orthogonally, without moving the attacker. If no enemy is in reach the action SHALL fail and cost no point.

#### Scenario: Ray at nearest
- **WHEN** the Wizard casts Ray with at least one enemy within 3 tiles
- **THEN** the nearest such enemy is attacked and the attacker does not move

#### Scenario: No target in reach
- **WHEN** Ray is cast with no enemy within 3 tiles
- **THEN** the cast is rejected, no point is spent, and the log states why

### Requirement: Experience drives levels

Kills SHALL award experience points, and crossing the cumulative level threshold SHALL immediately raise the character's level. A level increase SHALL raise level-based proficiency, grant the class's hit point gain, and update the experience display.

#### Scenario: Threshold crossing
- **WHEN** accumulated experience reaches the next cumulative threshold
- **THEN** the level rises once, bonuses used by checks and attacks reflect the new level, and hit point maximum increases by the class gain

#### Scenario: Kill awarding
- **WHEN** an enemy dies
- **THEN** its experience value is added to the running total

### Requirement: Enemies act with the same economy

After the player's turn ends, each living enemy SHALL act in turn with the same 3 action point budget, spending points to move toward the player and Strike when adjacent.

#### Scenario: Adjacent enemy strikes
- **WHEN** it is an enemy's turn and the player is adjacent
- **THEN** the enemy makes a Strike against the player resolved like a player Strike against AC

#### Scenario: Distant enemy closes
- **WHEN** it is an enemy's turn and the player is not adjacent
- **THEN** the enemy spends its points moving toward the player without attacking

### Requirement: Wizard casts a prepared spell

The Wizard's class skill verb SHALL cast the currently prepared spell. Cantrips cast without limit; a leveled spell consumes one slot of its level and SHALL be refused with a notice naming the slot level when none remain. A cast SHALL spend actions and slots only when it resolves against at least one foe: refusals for missing targets, missing actions, or missing slots leave those resources untouched. Preparing one of the Wizard's known spells is a free action.

#### Scenario: Cantrip casting
- **WHEN** the Wizard casts a cantrip with actions remaining and a foe in reach
- **THEN** the spell resolves and no spell slot changes

#### Scenario: Leveled spell spends a slot
- **WHEN** the Wizard casts Magic Missile while a first-level slot remains
- **THEN** the spell resolves and one first-level slot is consumed

#### Scenario: No slots left
- **WHEN** the Wizard casts a leveled spell whose slots of that level are exhausted
- **THEN** the cast is refused with a slot-exhausted notice and no foe takes damage

#### Scenario: Preparing is free and validated
- **WHEN** a prepare request names a known spell, an unknown spell, or the already prepared spell
- **THEN** known spells become the cast target without spending actions, while unknown spells leave the prepared spell unchanged

### Requirement: Spell effects follow their defined patterns

Each spell SHALL resolve by the pattern written in its spellbook entry. Aimed spells roll INT spell attacks with proficiency against AC, doubling the damage dice on a critical success, and stop at walls; blast spells deal their dice automatically with no roll. All spells cost one action except Fire Burst, which costs two.

#### Scenario: Ray and arc aim with spell attacks
- **WHEN** Ray of Frost or Electric Arc is cast at foes
- **THEN** each targeted foe faces an INT spell attack against its AC, the arc picking up to its two nearest foes within distance 2 for 1d4 + level + INT lightning each, heightening to 2d4 at character level 3

#### Scenario: Missiles never miss
- **WHEN** Magic Missile is cast at a foe within distance 5 regardless of the d20 result
- **THEN** no attack roll is made and the foe takes 1d4+1 force per bolt, with 1 + level/2 bolts

#### Scenario: Area blasts catch every foe in reach
- **WHEN** Burning Hands or Fire Burst is cast with multiple foes nearby
- **THEN** every foe within its radius — adjacent for Burning Hands, distance 2 for Fire Burst — takes its fire dice (2d6 and 4d6 respectively) with no roll while foes outside are untouched

#### Scenario: Fire Burst costs two actions
- **WHEN** Fire Burst resolves
- **THEN** two actions are spent, unlike the other four spells which cost one

### Requirement: The Wizard grimoire grows with level

At level 1 the Wizard's book holds the Ray of Frost and Electric Arc cantrips with no leveled slots. Reaching level 2 SHALL grant Magic Missile and one first-level slot; reaching level 3 SHALL grant Burning Hands and Fire Burst and raise the totals to two first-level and one second-level slot. Grants apply automatically whenever the Wizard's level rises and announce each learned spell in the log; other classes hold no spells at any level.

#### Scenario: Starting lineup
- **WHEN** a level 1 Wizard is created
- **THEN** the book contains Ray of Frost and Electric Arc, Ray of Frost is prepared, and no spell slots exist

#### Scenario: Level-up grants
- **WHEN** a Wizard gains the level that unlocks new spells
- **THEN** those spells appear in the book, slot totals rise, and the log lists each newly learned spell
