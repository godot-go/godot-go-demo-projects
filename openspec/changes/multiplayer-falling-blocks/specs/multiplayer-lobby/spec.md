## Purpose

Defines how players create and enter a shared game session, see who is present, signal readiness, and begin a match together — the pre-game lobby experience over the network.

## ADDED Requirements

### Requirement: Host a session
The system SHALL allow a player to host a session on a chosen port (defaulting to 7777), becoming the host, and SHALL show the hosting player in the lobby's player list with an address/port visible for others to use.

#### Scenario: Successful host
- **WHEN** a player chooses "Host" with port 7777 and an optional display name
- **THEN** the player enters the lobby screen, is listed in the roster, and other players on the same network can connect to that address and port

#### Scenario: Port in use
- **WHEN** hosting fails because the port is unavailable
- **THEN** the system shows an error and leaves the player on the title screen, able to retry

### Requirement: Join a session
The system SHALL allow a player to join a hosted session by entering the host's address and port, with a default of localhost:7777. On success the player enters the lobby; on failure the system shows an error and keeps the player on the join form.

#### Scenario: Successful join
- **WHEN** a player joins a reachable hosted session with a display name
- **THEN** the player sees the lobby and all existing players, and the existing players see the new player in their roster

#### Scenario: Unreachable host
- **WHEN** a player attempts to join an address with no listening host
- **THEN** the system displays a connection error and the player remains able to retry or return to the title

### Requirement: Session size cap
The host SHALL admit at most four players including itself and SHALL refuse new connections once the lobby is full; a refused player SHALL see an error and remain able to retry or return to the title.

#### Scenario: Fifth player refused
- **WHEN** a fifth player attempts to join a session that already has four players
- **THEN** the connection is refused with a visible error and the roster still lists exactly four players

### Requirement: Player roster
The lobby SHALL display, to every connected player, a list of all current players showing each player's display name, host designation, and connection/ready state. The roster SHALL update for all players when someone joins or leaves.

#### Scenario: Joiner appears in all rosters
- **WHEN** a third player joins a two-player lobby
- **THEN** all three players' rosters update within one second to show three entries with the new player's name

#### Scenario: Departure reflected
- **WHEN** a player disconnects from the lobby
- **THEN** the remaining players' rosters no longer list that player

### Requirement: Ready toggle
Every non-host player SHALL be able to toggle their own ready state, and the lobby SHALL show that state to all players. The host SHALL always be considered ready.

#### Scenario: Toggle visible to all
- **WHEN** a client marks itself ready
- **THEN** the client's entry shows as ready on every player's screen

### Requirement: Match start handshake
The host SHALL have a start control that is disabled until there are at least two connected players and all players are ready. When the host starts the match, every player's client SHALL transition to the versus match screen.

#### Scenario: Start gated on readiness
- **WHEN** the lobby has two players and one is not ready
- **THEN** the host's start control is unavailable

#### Scenario: Everyone advances
- **WHEN** the host starts with all players ready
- **THEN** all players enter the match within one second and a countdown or immediate spawn synchronizes play

### Requirement: Leaving the lobby
A player who leaves the lobby (or loses connection) SHALL return to the title screen, and the session SHALL remain usable by the others.

#### Scenario: Graceful leave
- **WHEN** a client leaves the lobby before the match starts
- **THEN** that client is at the title screen and the remaining players' rosters no longer list them
