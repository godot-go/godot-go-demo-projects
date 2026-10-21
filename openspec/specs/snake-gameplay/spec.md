# snake-gameplay Specification

## Purpose

Defines the rules and observable behavior of the snake demo: timed grid movement,
growth by eating pellets, loss on wall or self contact, and a run that is won at a
fixed pellet target.

## Requirements

### Requirement: Snake advances one cell per tick

The snake SHALL advance exactly one cell in its current direction each tick, and ticks SHALL be driven by a fixed timer rather than per-frame elapsed time so that movement speed does not depend on frame rate.

#### Scenario: Advance one cell per tick
- **WHEN** a tick elapses while the run is playing
- **THEN** the head moves one cell in the current direction and every body segment follows the segment ahead of it

#### Scenario: No movement between ticks
- **WHEN** time passes without a tick elapsing
- **THEN** the snake occupies the same cells as before

### Requirement: Turns apply on the next tick and reject reversal

A direction change SHALL take effect on the next tick. At most one turn SHALL be applied per tick, and a turn opposing the snake's current direction SHALL be ignored.

#### Scenario: Queued turn applies at the next tick
- **WHEN** the player turns while the snake is between ticks
- **THEN** the snake changes direction on the next tick, not immediately

#### Scenario: Reversal is ignored
- **WHEN** the player requests the direction opposite to the snake's current direction
- **THEN** the request is discarded and the snake continues in its current direction

#### Scenario: Only one turn per tick
- **WHEN** the player requests two perpendicular turns before the next tick
- **THEN** only the first request is applied and the snake does not turn twice in one tick

### Requirement: Eating a pellet grows the snake

When the head enters the pellet cell, the snake SHALL grow by one segment, the consumed pellet count SHALL increase by one, and a new pellet SHALL be placed.

#### Scenario: Growth on consumption
- **WHEN** the head moves onto the cell occupied by the pellet
- **THEN** the snake's length increases by one segment and its tail does not vacate

#### Scenario: Pellet counter advances
- **WHEN** the head consumes a pellet
- **THEN** the displayed consumed count increases by one and a new pellet becomes visible on the board

### Requirement: Pellets spawn on unoccupied cells

A new pellet SHALL occupy a free board cell that is not covered by any part of the snake.

#### Scenario: Spawn avoids the snake
- **WHEN** a pellet is placed
- **THEN** its cell is inside the board and contains no snake segment

### Requirement: The run is won at the pellet target

The run SHALL be won when the consumed pellet count reaches 20. After winning, the snake SHALL stop moving and the board SHALL remain visible until the player restarts or quits.

#### Scenario: Reaching the target wins
- **WHEN** the snake consumes the twentieth pellet
- **THEN** the run enters the won state and reports that the game was beaten

#### Scenario: Winning halts movement
- **WHEN** a tick elapses while the run is in the won state
- **THEN** the snake does not advance

### Requirement: Wall contact loses the run

The run SHALL be lost when the head would move outside the playfield. The snake SHALL NOT wrap to the opposite edge.

#### Scenario: Moving into a wall
- **WHEN** the head's next cell lies beyond the board edge
- **THEN** the run enters the lost state and reports the wall collision

### Requirement: Self contact loses the run

The run SHALL be lost when the head's next cell is occupied by the snake's own body, including the tail cell that would vacate on that same tick.

#### Scenario: Moving into the body
- **WHEN** the head's next cell contains any snake segment
- **THEN** the run enters the lost state and reports the collision

### Requirement: A run waits for the player's first input

A new run SHALL begin in a waiting state: the board and snake SHALL be visible, ticks SHALL NOT advance the snake, and a start prompt SHALL be shown. The run SHALL start when the player presses any key or mouse button. If that input is a direction action, that direction SHALL also apply on the first tick.

#### Scenario: Launch waits for input
- **WHEN** the game opens
- **THEN** the snake is drawn but does not move, and the start prompt is visible

#### Scenario: First input starts the run
- **WHEN** the player presses any key or mouse button while the run is waiting
- **THEN** the prompt disappears and ticks begin

#### Scenario: Starting press steers
- **WHEN** the starting input is a direction action
- **THEN** the run starts and the snake turns that way on its first tick

#### Scenario: Restart waits for input
- **WHEN** the player restarts after a win or loss
- **THEN** the run resets to length 3, pointed right, with zero pellets consumed, and waits for the player's first input

### Requirement: Run state is displayed and restartable

The game SHALL display pellet progress toward the target and SHALL show a distinct message when the run is won or lost. Restart SHALL begin a fresh run with length 3, zero pellets consumed, and movement to the right. Pause SHALL stop ticks until resumed.

#### Scenario: Progress is visible during play
- **WHEN** the run is playing
- **THEN** the interface shows consumed pellets against the 20-pellet target

#### Scenario: Restart after a loss
- **WHEN** the player restarts from the lost state
- **THEN** the snake starts at length 3, pointed right, with the pellet count reset, waiting for the player's first input

#### Scenario: Pause halts the run
- **WHEN** the player pauses during play
- **THEN** ticks stop affecting the snake until the player resumes
