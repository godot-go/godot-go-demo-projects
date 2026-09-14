## Purpose

Defines online competitive play built on the lobby: each player runs their own game under the falling-blocks gameplay rules while line clears send garbage to opponents, and the session resolves a winner.

## ADDED Requirements

### Requirement: Synchronized match start
The system SHALL start every connected player's game (empty board, fresh piece sequence, zero score) at the same signal from the host.

#### Scenario: Fair start
- **WHEN** the host signals match start and all clients begin
- **THEN** every player's board is empty and gravity begins on the start signal

### Requirement: Garbage line attacks
When a player locks a piece that clears lines, the system SHALL send garbage to every other in-match player: 0 lines for a single, 1 for a double, 2 for a triple, 4 for a quad, plus 1 extra per consecutive clear combo. Received garbage SHALL be inserted as a full row (with one random gap) beneath the recipient's stack when the recipient's next piece locks, pushing existing cells up.

#### Scenario: Quad sends four
- **WHEN** a player clears four rows in one lock with no prior combo
- **THEN** each opponent is queued to receive 4 garbage rows

#### Scenario: Garbage applies on next lock
- **WHEN** a player with 2 queued garbage rows locks their next piece
- **THEN** 2 rows with gaps are inserted at the bottom, the stack rises by 2, and the queue empties

#### Scenario: Top-out from garbage
- **WHEN** inserted garbage pushes cells into the spawn area
- **THEN** the affected player's game is over per the game-over rules

### Requirement: Opponent status display
Each player SHALL see, for every opponent: name, score, total lines, pending incoming garbage count, and alive/eliminated state, updated within one second of the opponent's change. Opponents' full board layouts are not required.

#### Scenario: Opponent score visible
- **WHEN** an opponent clears a quad
- **THEN** that opponent's score and lines increment on every other player's screen within one second

### Requirement: Win determination
The system SHALL eliminate a player when their game ends and declare the last remaining player the winner. In a two-player match the first top-out decides the loser. Results SHALL be shown on all clients.

#### Scenario: Last player standing
- **WHEN** every player but one has topped out
- **THEN** the remaining player is shown as the winner and the match stops accepting play

#### Scenario: First to top out loses a duel
- **WHEN** in a two-player match one player tops out
- **THEN** the other player is shown as winner and both boards stop

### Requirement: Rematch
The match results screen SHALL offer the host a rematch action that returns all players to a synchronized match start with reset scores, boards, and garbage queues.

#### Scenario: Rematch resets state
- **WHEN** the host triggers rematch from the results screen
- **THEN** all players are back in play with zero score, empty boards, and no queued garbage

### Requirement: Disconnect during match
If a player disconnects mid-match, the system SHALL mark them eliminated for win purposes and continue the match for remaining players; if fewer than two players remain, the match SHALL end with the remaining player as winner (or a draw if none remain).

#### Scenario: Opponent quits mid-match
- **WHEN** one of three players disconnects during a match
- **THEN** the other two continue, and the disconnected player counts as eliminated for win determination
