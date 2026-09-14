# falling-blocks-gameplay Specification

## Purpose

Defines the core falling-blocks rules the demo must implement: piece generation, player control of falling pieces, gravity, line clearing, scoring, leveling, and game over, independent of any networking.

## Requirements

### Requirement: Tetromino set and generation
The system SHALL cycle the seven standard tetrominoes (I, O, T, S, Z, J, L) using a bag randomizer: every consecutive block of seven spawned pieces contains each tetromino exactly once, in random order. The system SHALL display at least the next upcoming piece in a preview area.

#### Scenario: Bag completeness
- **WHEN** ten pieces have spawned from a fresh game
- **THEN** the first seven spawned contain all seven distinct tetrominoes, and the eighth begins a new bag

#### Scenario: Next-piece preview
- **WHEN** any piece is the active falling piece
- **THEN** the preview shows the identity of the piece that will spawn after the current one locks

### Requirement: Piece movement and rotation
The system SHALL let the player move the active piece one cell left or right and rotate it clockwise and counter-clockwise. A move or rotation SHALL be rejected if no orientation of the piece fits inside the playfield bounds and empty cells; rotation SHALL attempt horizontal kick offsets of up to one cell in each direction (up to two cells for the I piece) before rejecting.

#### Scenario: Move against wall
- **WHEN** the player presses left while the piece touches the left wall
- **THEN** the piece does not move and no cells change

#### Scenario: Wall kick on rotate
- **WHEN** the player rotates a T piece flat against the left wall such that rotation would place cells outside the wall
- **THEN** the piece rotates shifted one cell right, or the rotation is rejected with the piece unchanged

#### Scenario: Rotation into occupied cell
- **WHEN** a rotation would overlap any locked cell
- **THEN** the rotation (and its kicks) is rejected and the piece is unchanged

### Requirement: Soft drop, hard drop, and locking
The system SHALL support soft drop (piece moves down one cell per input with increased gravity rate and 1 point per entered cell) and hard drop (piece teleports to the lowest valid position, locks immediately, and awards 2 points per entered cell). A piece SHALL also lock when it cannot fall further under gravity.

#### Scenario: Hard drop locks immediately
- **WHEN** the player triggers hard drop
- **THEN** the piece moves to the lowest empty position in its column, is locked into the stack, and a new piece spawns

#### Scenario: Soft drop scoring
- **WHEN** the player soft drops a piece three cells before it locks naturally
- **THEN** the score increases by 3 and the piece occupies the lowered cells

### Requirement: Gravity
The system SHALL move the active piece down one cell at a fixed interval determined by the current level; the piece SHALL not lock until the next gravity step would collide.

#### Scenario: Idle piece descends
- **WHEN** no input is received for one gravity interval
- **THEN** the active piece occupies the cell one row lower

### Requirement: Line clearing and scoring
When the active piece locks, the system SHALL remove every completely filled row, shift all rows above down, and score per simultaneously cleared rows: 1=100, 2=300, 3=500, 4=800, multiplied by the current level.

#### Scenario: Double clear scoring at level 2
- **WHEN** locking a piece completes exactly two rows while the level is 2
- **THEN** both rows are removed, the stack shifts down, and the score increases by 600

#### Scenario: Partial row not cleared
- **WHEN** locking a piece leaves any gap in a row
- **THEN** no rows are removed

### Requirement: Level progression
The system SHALL increase the level by one for every 10 cumulative lines cleared and increase gravity speed accordingly.

#### Scenario: Level up at 10 lines
- **WHEN** the 10th line of the game is cleared
- **THEN** the displayed level increases from 1 to 2 and gravity interval shortens

### Requirement: Game over
The system SHALL end the player's game when a newly spawned piece cannot occupy its spawn cells without overlapping locked cells or leaving the playfield, or when any cell locked into the stack occupies a hidden buffer row above the visible playfield. Upon game over the final score SHALL be shown and no further inputs SHALL affect the board.

#### Scenario: Top-out ends the game
- **WHEN** the stack reaches the spawn area and a new piece would overlap it
- **THEN** the game is marked over for that player and the final score is displayed

### Requirement: Pause
The system SHALL allow the player to pause and resume a solo (offline) game; while paused, gravity, input, and timers SHALL be frozen. Online versus matches SHALL NOT offer pausing.

#### Scenario: Paused game freezes
- **WHEN** the player pauses and waits longer than the gravity interval
- **THEN** the active piece has not moved, and after resume normal play continues

### Requirement: Ghost landing preview
The system SHALL expose and render the cells the active piece would occupy if hard-dropped immediately, as a static landing shadow distinct from the active piece. The ghost SHALL update as the piece moves or rotates and SHALL not affect locking, scoring, or the board.

#### Scenario: Ghost tracks the active piece
- **WHEN** the active piece is above an empty column
- **THEN** the ghost cells appear on the floor and shift horizontally with the piece

#### Scenario: Ghost rests on the stack
- **WHEN** the active piece is above filled rows
- **THEN** the ghost cells sit directly on top of the highest occupied cells in the piece's columns

### Requirement: Held-input auto-shift
The system SHALL treat a fresh left/right press as a single one-cell move and, while the key is held, repeat horizontal moves only after an initial delay and then at a faster repeat interval. Changing direction during a hold SHALL immediately move one cell in the new direction and restart the delay; releasing the key SHALL reset the repeater. Holding the soft-drop key SHALL repeat the one-cell drop at a fixed short interval while pressed.

#### Scenario: Single tap moves once
- **WHEN** the player presses and releases left without holding
- **THEN** the piece moves exactly one cell left and no further movement occurs

#### Scenario: Hold repeats after delay
- **WHEN** the player holds right through the initial delay window
- **THEN** the piece stays still during the delay, then moves repeatedly at the fast repeat rate until released

### Requirement: Solo mode
The system SHALL offer an offline single-player game starting from the title screen that applies the falling-blocks gameplay rules without any networking. On game over the final score SHALL be shown and the player SHALL be able to start another game.

#### Scenario: Play without a network
- **WHEN** the player chooses Solo while hosting and joining nothing
- **THEN** a match begins immediately under the standard rules and no multiplayer peer is created
