# Spec Delta

## Purpose

Defines the boot flow before the town: a start screen offering New Game, Continue, and Quit; a four-step-free character generation path for New Game (played-save confirm, class roster, flavor card); and the persistence moment that turns the choice into a save.

## ADDED Requirements

### Requirement: Start screen offers New Game, Continue, and Quit

The game SHALL boot to a start screen instead of the town, rendered as full-screen key-driven text pages in the existing center-panel idiom. It SHALL offer exactly three entries — New Game, Continue, Quit — navigable with the move keys and confirmed with the same keys the gate roster uses (number shortcuts plus Enter). Quit SHALL exit the application cleanly; the window close button SHALL keep working.

#### Scenario: Fresh boot shows the menu
- **WHEN** the game starts
- **THEN** the menu lists New Game, Continue, and Quit, with Up/Down (and number keys) moving the selection and Enter activating nothing until pressed

#### Scenario: Quit from the menu
- **WHEN** the player confirms Quit
- **THEN** the application exits with no error output

#### Scenario: No other verb leaks
- **WHEN** any other key is pressed on a start-screen page
- **THEN** it has no effect on town or run state (none exists yet)

### Requirement: Continue loads the banked save or stands down

Continue SHALL load the banked meta (renown, building levels, selected class) into the town — the same state a played save keeps between sessions. When the save file is missing or reads fresh (no renown, no levels, no first-kills), Continue SHALL render unselectable with a "no save yet" hint rather than silently resetting.

#### Scenario: Continue with a played save
- **WHEN** a save with banked progress exists and Continue is confirmed
- **THEN** the town loads with that renown and building state and the selected class

#### Scenario: Continue with nothing banked
- **WHEN** the save file is missing or contains only fresh-install defaults
- **THEN** the Continue row reads "no save yet" and confirming it does nothing

### Requirement: New Game resets a played save only through confirmation

Picking New Game when a played save exists SHALL show a confirmation page naming what is lost (renown, building levels, first-kills) before resetting meta to fresh-install defaults; Esc SHALL back out untouched. With no played save the confirm SHALL be skipped. The old save SHALL be considered gone the moment the player confirms and the character generation completes.

#### Scenario: Confirm guards a played town
- **WHEN** New Game is confirmed and a played save exists
- **THEN** the confirm page appears; Esc returns to the menu with the save untouched

#### Scenario: Fresh install skips confirm
- **WHEN** New Game is confirmed with no played save
- **THEN** character generation begins immediately

### Requirement: Character generation offers the unlocked-class roster

Character generation SHALL list exactly the classes unlocked by Guild levels (matching the gate roster), each with its stat-line blurb from the shared class tables, and confirming SHALL set that class on the meta.

#### Scenario: Roster mirrors the gate
- **WHEN** character generation begins on a fresh meta
- **THEN** only the starter class is offered, and richer rosters appear as Guilds are bought

### Requirement: Character generation rolls a flavor card

On class confirm, character generation SHALL present a deterministic flavor card rolled from a seed — a hometown line, a boon line, four rolled tale numbers, and a sigil color tint — as pure flavor with no mechanical stat effect. Esc SHALL reroll the card with a fresh seed; confirming SHALL persist the flavor and class into the save immediately and enter the town. The character sheet SHALL later show the same stored card.

#### Scenario: Card is deterministic but rerollable
- **WHEN** the same seed is rolled twice
- **THEN** the hometown, boon, tale rolls, and sigil are identical; Esc rolls a new seed

#### Scenario: Accept persists
- **WHEN** the player confirms the card
- **THEN** the save file holds the selected class and flavor, and the town opens with the usual welcome
