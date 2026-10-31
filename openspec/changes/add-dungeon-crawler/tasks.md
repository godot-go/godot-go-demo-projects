# Tasks

## 1. Module scaffold

- [x] 1.1 Create `2d/dungeon_crawler` Go module (`go.mod`, `main.go` gdextension entry, `Makefile` adapted from `2d/snake`) and register it in `go.work`; verify `make build` emits `project/lib/libgodotgo-2ddungeoncrawler-linux-amd64.so`
- [x] 1.2 Add the Godot project skeleton (`project.godot`, `dungeoncrawler.gdextension`, empty `Main.tscn` with the registered root class); verify `godot --headless --path project/ --editor --quit` loads the extension with no errors
- [x] 1.3 Write `2d/dungeon_crawler/README.md` following the other demos and verify its build/run commands match the Makefile targets

## 2. Rules core (pure Go, `pkg/crawler`)

- [x] 2.1 Define grid, tile, entity, actor, and visibility types with a headless unit test for occupancy and movement basics
- [x] 2.2 Implement the d20 check primitive (trained = level + 2 proficiency, ability modifier, natural 20/1 degrees, untrained-only-trained rejection) with seeded-`rand` unit tests
- [x] 2.3 Add `numbers.go` with PF2e XP thresholds, class ability arrays, monster statline generator, DC tiers, and renown payout weights, plus unit tests pinning pacing invariants (e.g. a floor-1 clear reaches level 2)

## 3. Combat engine

- [x] 3.1 Implement the 3-action turn (Stride, bump Strike, multiple attack penalty −5/−10, auto-end at 0, early end) with unit tests
- [x] 3.2 Implement classes as data (Fighter/Cleric/Wizard/Rogue), their combat skills (Power Attack, Lay on Hands, Ray with nearest-in-reach-3 targeting, Sneak Attack condition) as action-costed verbs with unit tests
- [x] 3.3 Implement XP award on kill and immediate level-up applying proficiency and HP gains, with unit tests including threshold edges
- [x] 3.4 Implement enemy activation after the player's turn (shared 3-action economy, stride-toward/strike-when-adjacent policy) with unit tests for both behaviors
- [x] 3.5 Document the engine verb/event API in package doc comments and verify `go vet ./...` and tests pass

## 4. Dungeon generation and visibility

- [x] 4.1 Implement seeded BSP room-and-corridor generation with entrance, stairs, and room typing; verify by unit-testing reachability (stairs and every room reachable from entrance) across many seeds
- [x] 4.2 Implement per-floor feature placement (locked doors, traps, hidden doors, potions, monster spawns from floor-tier tables, floor-5 boss) with unit tests asserting tier correctness and feature presence
- [x] 4.3 Implement symmetric shadowcasting FOV with explored-memory bitmask and visibility-only entity export; verify with unit tests for unseen/explored/visible states

## 5. Run orchestration

- [x] 5.1 Implement run lifecycle (`NewRun(seed)`, floor descent, death, boss-kill win, end-state freeze) emitting structured events; verify with unit tests covering win, death, and descend flows
- [x] 5.2 Implement feature verbs (Athletics force, Thievery disable, trap trigger, Search with Archive range, potion pickup/quaff) with unit tests per feature
- [x] 5.3 Implement renown payout (kills by level, floors, first-kill bonus) and end-screen payload; verify with unit tests for death and victory payouts

## 6. Town and meta

- [x] 6.1 Implement the town map model with bump-interactable buildings, gate, and guild class-selection state; verify with model-level unit tests
- [x] 6.2 Implement upgrade definitions and purchase rules (level caps, rising costs, insufficient-renown rejection) with unit tests
- [x] 6.3 Implement JSON meta save/load (renown, building levels, first-kills) with unit tests for round-trip and fresh-install defaults
- [x] 6.4 Implement run-start application of upgrade effects (Smithy, Clinic, Library, Archive, Guild unlock + level-2 enhancement) with unit tests

## 7. Godot shell

- [x] 7.1 Wire map rendering: TileMapLayer cells and fog modulate repaint from engine snapshots (adapt `2d/topdown` tileset); verify by launching a seeded floor and observing layout and fog states
- [x] 7.2 Wire input (WASD/arrows move+bump, Space class skill, E Search, panel confirm) to engine verbs and render the scrolling combat log; verify observable behavior matches verbs
- [x] 7.3 Add the HUD (HP bar, 3 action pips, XP bar, floor number, carried potions) and verify it tracks engine state after actions
- [x] 7.4 Build town mode rendering plus building panels and class select, and wire `user://` save path through the shell; verify a purchase persists across relaunch
- [x] 7.5 Add win/loss end screens with renown payout and run seed; verify both states are reachable and display correctly

## 8. Integration and balance

- [x] 8.1 Add a headless scripted-bot playthrough test (walk, fight, descend, die or win) verifying a full run reaches an end state
- [x] 8.2 Playtest passes on floors 1–5, tune `numbers.go` for pacing, and verify the invariant tests still pass
- [x] 8.3 Verify `go test ./...` across the workspace is green and `make build` + `make run` work from a clean tree

## 9. Wizard spells and the spellbook

- [x] 9.1 Model spells, per-level grants, and slot totals in `spells.go` with `syncWizardBook` / `refreshWizardSlots` hooks
- [x] 9.2 Route the Wizard class skill verb through the prepared-spell dispatcher, keeping Ray of Frost's existing resolution
- [x] 9.3 Implement Electric Arc, Magic Missile, Burning Hands, and Fire Burst effects
- [x] 9.4 Hook level-up grants into kill awards and slot refresh into descent
- [x] 9.5 Add the spellbook section to the character sheet
- [x] 9.6 Build the pausing spellbook modal (B) with prepare keys 1-5, help text, and hotkey footer
- [x] 9.7 Green the rules suite (`spells_test.go`) and a headless E2E driver over the modal flow

## 10. Animated avatars

- [x] 10.1 Add pure idle-motion math (`bob`, `blinkClosed`, `pulse01`, `sweep01`) in `idleanim.go` with unit tests pinning ranges, blink windows, and glint cycles
- [x] 10.2 Add polygon and disc painting primitives to `ui.go` (`drawPolygon`, `drawDisc`) over the `draw_colored_polygon` path
- [x] 10.3 Implement `avatars.go` painters: class figures with signature props (sword glint, orbiting halo, swaying hat and staff, hooded dagger flash), nine bestiary idles (tail flick, ear twitch, bone rattle, breathing, spectral float, orb glow, torn hem, boss aura and crown), and item figures (sloshing potion, blade shine, stair shimmer, straining door, chomping trap jaws)
- [x] 10.4 Wire avatars into run and town rendering with per-entity `fxPhase` desync; explored-unseen items freeze at `t=0` dimmed; `advanceFx` repaints every frame
- [x] 10.5 Verify: vet and rules tests green, `make build`, and a windowed E2E driver (town wanderer, Wizard run start, hundreds of animated frames) with zero engine errors or warnings

## 11. Rogues pick locks

- [x] 11.1 Rework the locked-door bump to attempt the player's trained world skill — Athletics forces (Fighter), Thievery picks (Rogue), neither means still locked — with per-skill success/failure log flavor; pin with `TestBumpLockedDoorPickedByRogue` and `TestRoguePickFailureKeepsDoorLockedAndSpends`

## 12. Vault locks and Knock

- [x] 12.1 Seal the locked door into a vault: `findVault` picks an away room, walls surplus doorways, locks the border doorway (never room interior), verifies floor-wide connectivity (stairs intact, interior sealed, nothing severed) with rollback per candidate; loot — the potion always, the weapon on half the floors — lives inside, and traps/monsters never spawn within
- [x] 12.2 Pin the contract across seeds and floors (1–4 with stairs reach, 5 without) in `vault_test.go` (`TestLockedDoorSealsVaultDoorway`)
- [x] 12.3 Add Knock (Wizard level 2 cantrip, 1 action, no slot): opens one locked door within a step, fizzle commits nothing; spell-count pins in `TestSpellGrantsByLevel` moved 5 → 6

## 13. Town facelift

- [x] 13.1 Pure painter helpers in `pkg/demo/town.go` — `condition01` kept-ness ladder, `tile01` scatter hash, `townWear` over a new `Meta.UpgradeTotal` — pinned by `town_test.go` for bounds, monotonic climb, and wear
- [x] 13.2 Rewire `drawTown` into three passes: ground (mud→grass with whole-town wear, moss/stains/weeds/tufts/flowers per tile), wander trail (ruts→pebbled path), then facades; keep class avatar on top
- [x] 13.3 Implement the facade ladder in `drawTownFacade`: leaning roofs → missing shingles/moss/cracks/shattered-or-boarded windows/unlatched doors/rubble/weeds, patched brightening, capped trim + flowerboxes + chimney smoke + pulsing lantern + doormat; gate rides town wear with torches → battlements
- [x] 13.4 Windowed E2E smokes over a derelict and a fully-kept town, each idling 120 frames with zero engine errors or warnings; driver removed after

## 14. Spells destroy traps

- [x] 14.1 Add `trapsWithin` reach-geometry helper and aimed-spell fallbacks — Ray of Frost, Electric Arc, Magic Missile target the nearest armed trap (same reach + line-of-sight as foe targeting) only when no foe is in reach; foe priority and no-target fizzle pinned by tests
- [x] 14.2 Catch armed traps in the fire-blast template: Burning Hands and Fire Burst destroy traps in radius alongside foes, spending the cast once for anything hit
- [x] 14.3 One hit destroys — no roll, no trap AC; the trap flips to the Rogue-disarmed state; spec delta "Spells destroy armed traps" with scenarios and `openspec validate`

## 15. Pixel-art town

- [x] 15.1 Sprite engine in `pkg/demo/pixelart.go` — `sprite` grids rasterized by `spriteCells` (transparent skip, order pinned, ragged panic), a story ink palette, per-kind `buildingPal` roof hues, and run-length `blit` through the polygon path; rasterization, palette coverage, and roof hue dominance pinned by `pixelart_test.go`
- [x] 15.2 Generated facades + gate sprite (cottage gable, eaves, windows, sign glyphs per kind; 24-wide gatehouse with arch and banner) and prop sprites (tree, pine, well, stall, crate, barrel, lamp)
- [x] 15.3 Decay tiers as sprite edits — ruin punches roof holes, boards one window, shatters the other, cracks plaster and dims the palette; patched lights a pane; kept adds awning stripes, flowerboxes, lantern, chimney; cache per building/tier, glows animated off `c.now`
- [x] 15.4 Ground/trail/plaza pixel painters with the wear ladder + hash dither, `townProps` prop layout (well plaza, Smithy market clutter, gate lamps, tree/pine scatter off buildings/trail/spawn) glued to `drawTown` pass order; determinism, walkable-avoidance, and no-overlap pinned by tests
- [x] 15.5 Faster ink path — `drawFilledRect` routed to `RenderingServer.canvas_item_add_rect` with cached node RID (was per-call PackedVector2Array churn); windowed E2E smokes on derelict and fully-kept towns both PASS with zero engine errors or warnings

## 16. Pixel-art dungeon ground

- [x] 16.1 Run-art core in `pkg/demo/dungeonart.go` — 7x7 art grids at 3px per cell RLE-merged into `artRun` rectangles over a dungeon palette (brick/mortar/lit facing, slab/joint/speckle over the story battery); coverage, palette resolve, bond, slab joints, determinism, and spatial variance pinned by `dungeonart_test.go`
- [x] 16.2 Hash decor inside the pattern — wall moss clumps and hairline cracks only within brick courses, floor scuff patches plus rare rubble, bone, or slab cracks (one per tile) via the seeded hash; decor never covers pattern joints
- [x] 16.3 `drawRun` rewired onto `drawArtTile` (one fill per run, dim through the pen) for walls and floors; feature cells keep their flat tint, hidden doors stay wall, avatar loops untouched; windowed E2E run smoke PASS with zero engine errors or warnings

## 17. Start screen and character generation

- [x] 17.1 Flavored character in `pkg/crawler` — `RollFlavor(seed)` returning a deterministic `CharacterFlavor` (hometown, boon, four tale rolls, sigil index) from hand-written tables, `Meta.Flavor` persisted with nil-tolerant load; determinism, table bounds, and round-trip pinned by tests (RED first)
- [x] 17.2 Start-screen pages in `pkg/demo` — `modeStart` with menu (New Game / Continue / Quit), played-save confirm page, class roster page off `UnlockedClasses()`, and flavor-card page with Esc reroll; key-driven like the gate roster, no verb leakage, Quit exits cleanly
- [x] 17.3 Continue gating and persistence — save-file presence + fresh-save detection greys Continue; accepting the card writes meta (class + flavor) immediately and enters town; character-sheet modal shows the stored card
- [x] 17.4 Windowed E2E smoke — fresh boot menu, Continue greyed, New Game confirm path (played save), class pick, reroll + accept, town reached with zero engine errors or warnings; save restored afterward

## 18. Trap spotting and hostile outlines

- [x] 18.1 Hidden traps in `pkg/crawler` — `Floor` spotted-trap set with `SpotTrap`/`TrapSpotted` (run-only state), `TrapSpotDC` constant, and `Search` extending its single Perception roll to reveal the nearest hidden thing (door or trap) with the usual roll/DC log; unspotted traps trigger as before; spotted set pinned by tests (RED first)
- [x] 18.2 Shell rendering — unspotted armed traps stop rendering; spotted traps draw (dimmed out of sight); every living enemy in `Visible` gets a thin hostile outline tile-frame that memory ghosts never carry; windowed E2E run smoke PASS with zero engine errors or warnings

## 19. Eight-direction movement

- [x] 19.1 Diagonal engine in `pkg/crawler` — four diagonal `Dir` values with `Delta` offsets pinned by table test; `Run.Move` strides diagonally for 1 action, bump-strikes diagonally, and blocks corner-clips (both orthogonal neighbours unwalkable) for free; `Town.Move` walks and bumps (gate included) on diagonals; tests RED first
- [x] 19.2 Shell input and text — `shiftHeld` helper, pure `diagonalFor` resolver (Shift + direction tilts one step clockwise: ↑↗ →↘ ↓↙ ←↖) pinned by tests, wired into `directionFromInput` with directions polled non-exact so Shift-modified presses reach the game; help/hotkey lines mention Shift diagonals; windowed E2E smoke injecting OS-realistic Shift+dir through the gate and diagonally in a run with zero engine errors or warnings

## Workflow follow-up

- Archive the change after review, letting the three capability specs land in `openspec/specs/`.
