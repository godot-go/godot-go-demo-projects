package demo

import (
	"fmt"
	"math"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// Damage feedback: every hit floats a number off its receiver, and a big
// hit (critical success or 10+ damage) shakes the receiver for a beat. All
// state advances on the game clock in _process(delta) and paints in _draw;
// the jitter is a decayed sine keyed per receiver, so it is deterministic
// and costs no RNG draws the rules could notice.

const (
	popupLife    = 0.9  // seconds a damage number floats
	popupRise    = 26.0 // pixels it travels while fading
	shakeLife    = 0.45 // seconds a shaken receiver rattles
	shakeAmp     = 3.5  // max pixel displacement, decays linearly to 0
	bigHitDamage = 10   // non-crit threshold that still counts as big
)

// popup is one floating damage number.
type popup struct {
	text      string
	x, y      float32
	col       rgb
	size      int32
	age, life float64
}

// fxSpan is a wall-clock window on the game clock, [start, until).
type fxSpan struct{ start, until float64 }

// spawnHitFx reacts to a batch of rules events: damage numbers pop over the
// tile where the blow landed and the receiver — the player or the struck
// enemy — starts shaking when the hit is big.
func (c *DungeonCrawler) spawnHitFx(evs []crawler.Event) {
	if c.mode != modeRun || c.run == nil {
		return
	}
	for _, ev := range evs {
		if ev.Damage <= 0 || ev.Subject == "" {
			continue
		}
		crit := ev.Degree == crawler.CriticalSuccess
		text := fmt.Sprintf("-%d", ev.Damage)
		size := int32(15)
		col := rgb{0.95, 0.93, 0.75} // pale gold on an enemy
		if ev.Subject == "player" {
			col = rgb{0.95, 0.4, 0.35} // red when it is you bleeding
		}
		if crit {
			text = fmt.Sprintf("%d!", ev.Damage)
			size = 22
			col = rgb{0.95, 0.82, 0.35} // bright gold for a crit, whoever took it
		}
		x, y := c.tileCenter(ev.At)
		c.popups = append(c.popups, popup{
			text: text, x: x, y: y - 4,
			col: col, size: size, life: popupLife,
		})
		if crit || ev.Damage >= bigHitDamage {
			key := "player"
			if ev.Subject == "enemy" {
				key = fmt.Sprintf("e%d", ev.EnemyID)
			}
			if c.shakes == nil {
				c.shakes = map[string]fxSpan{}
			}
			c.shakes[key] = fxSpan{start: c.now, until: c.now + shakeLife}
		}
	}
}

// advanceFx ticks the popup and shake clocks; while any effect is alive it
// asks for a repaint so the animation runs even with no input.
func (c *DungeonCrawler) advanceFx(delta float64) {
	if delta <= 0 {
		return
	}
	c.now += delta
	if len(c.popups) > 0 {
		kept := c.popups[:0]
		for _, p := range c.popups {
			p.age += delta
			if p.age < p.life {
				kept = append(kept, p)
			}
		}
		c.popups = kept
	}
	for k, w := range c.shakes {
		if c.now >= w.until {
			delete(c.shakes, k)
		}
	}
	// Avatars idle continuously, so the view repaints every frame.
	c.QueueRedraw()
}

// shakeOffset returns the current pixel jitter for a receiver key
// ("player" or "e<ID"); the wobble decays linearly over its window.
func (c *DungeonCrawler) shakeOffset(key string) (float32, float32) {
	w, ok := c.shakes[key]
	if !ok || c.now >= w.until || w.until <= w.start {
		return 0, 0
	}
	decay := float32((w.until - c.now) / (w.until - w.start))
	amp := shakeAmp * decay
	ph := fxPhase(key)
	dx := amp * float32(math.Sin(c.now*55.0+ph))
	dy := amp * float32(math.Sin(c.now*41.0+ph*1.7))
	return dx, dy
}

// fxPhase spreads the shake rhythm per receiver so two shaken enemies do
// not wobble in lockstep.
func fxPhase(key string) float64 {
	h := 0
	for i := 0; i < len(key); i++ {
		h = h*31 + int(key[i])
	}
	return float64(h%97) * 0.64
}

// tileCenter maps a map cell to the center of its tile in window space,
// matching drawRun's transform.
func (c *DungeonCrawler) tileCenter(cell crawler.Cell) (float32, float32) {
	ox := float32((winW - c.run.Floor.W*runTile) / 2)
	oy := float32(mapTop)
	return ox + float32(cell.X*runTile) + runTile/2, oy + float32(cell.Y*runTile) + runTile/2
}

// drawPopups paints every live damage number rising and fading above the
// map. The fallback font reference is released at the end of the frame
// rather than waiting on the finalizer, and every String created here is
// destroyed before returning.
func (c *DungeonCrawler) drawPopups() {
	if len(c.popups) == 0 {
		return
	}
	font := GetThemeDBSingleton().GetFallbackFont()
	defer font.Unref()
	if !font.IsValid() { // no glyphs to draw with (defensive; UI builds always have one)
		return
	}
	for _, p := range c.popups {
		t := float32(p.age / p.life)
		// Left-aligned with width -1, so pre-shift to center the text.
		approx := float32(len(p.text)) * float32(p.size) * 0.6
		x := p.x - approx/2
		y := p.y - popupRise*t
		col := NewColorWithFloat32Float32Float32Float32(p.col.r, p.col.g, p.col.b, 1-t)
		txt := NewStringWithUtf8Chars(p.text)
		pos := NewVector2WithFloat32Float32(x, y)
		c.DrawString(font, pos, txt,
			constant.HORIZONTAL_ALIGNMENT_LEFT, -1, p.size, col,
			constant.TEXT_SERVER_JUSTIFICATION_FLAG_JUSTIFICATION_NONE,
			constant.TEXT_SERVER_DIRECTION_DIRECTION_LTR,
			constant.TEXT_SERVER_ORIENTATION_ORIENTATION_HORIZONTAL, 1.0)
		txt.Destroy()
	}
}
