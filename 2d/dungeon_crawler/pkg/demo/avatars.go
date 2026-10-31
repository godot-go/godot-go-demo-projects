package demo

import (
	"math"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// Procedural avatars: every character and floor item is drawn from
// polygons and discs, posed by the idle math in idleanim.go instead of
// living as a flat colored box. Painters receive the tile box (x, y, s),
// a static phase (from fxPhase) that keeps neighbours out of lockstep, the
// shared clock t, and a dim flag — explored memory freezes at t=0 and
// fades every color, matching how floor tiles dim.

// pen wraps the crawler node with the dim decision and the three shape
// primitives, so sprite code reads like drawing, not marshalling.
type pen struct {
	c   *DungeonCrawler
	dim bool
}

func (p pen) col(v rgb) rgb {
	if p.dim {
		return v.dimmed()
	}
	return v
}

func (p pen) fill(x, y, w, h float32, v rgb) {
	drawFilledRect(p.c, x, y, w, h, p.col(v).color())
}

func (p pen) poly(pts [][2]float32, v rgb) {
	drawPolygon(p.c, pts, p.col(v).color())
}

func (p pen) disc(cx, cy, r float32, v rgb) {
	drawDisc(p.c, cx, cy, r, p.col(v).color())
}

// glow paints a translucent disc for auras, lights, and spectral bodies.
// Alpha is float64 so callers can feed pulse01/sweep01 output directly.
func (p pen) glow(cx, cy, r float32, v rgb, a float64) {
	col := p.col(v)
	if p.dim {
		a *= 0.4
	}
	drawDisc(p.c, cx, cy, r, NewColorWithFloat32Float32Float32Float32(col.r, col.g, col.b, float32(a)))
}

// polyGlow is glow for arbitrary shapes — the shadow's translucent body.
func (p pen) polyGlow(pts [][2]float32, v rgb, a float64) {
	col := p.col(v)
	if p.dim {
		a *= 0.4
	}
	drawPolygon(p.c, pts, NewColorWithFloat32Float32Float32Float32(col.r, col.g, col.b, float32(a)))
}

func shade(v rgb, f float32) rgb { return rgb{v.r * f, v.g * f, v.b * f} }

var (
	skin     = rgb{0.93, 0.76, 0.6}
	steel    = rgb{0.72, 0.75, 0.82}
	gold     = rgb{0.92, 0.78, 0.3}
	white    = rgb{0.96, 0.96, 0.98}
	darkEye  = rgb{0.08, 0.08, 0.12}
	classInk = map[string]rgb{
		"Fighter": {0.6, 0.64, 0.76},
		"Cleric":  {0.88, 0.85, 0.75},
		"Wizard":  {0.36, 0.4, 0.85},
		"Rogue":   {0.3, 0.48, 0.36},
	}
)

// drawClassAvatar paints the four classes as tiny figures: shared bob and
// blink, plus a signature prop each — sword, halo, hat and staff, hood.
func drawClassAvatar(c *DungeonCrawler, class string, x, y, s float32, phase, t float64, dim bool) {
	p := pen{c, dim}
	ink, ok := classInk[class]
	if !ok {
		ink = classInk["Fighter"]
	}
	by := bob(t, phase)
	cx := x + 0.5*s
	hy := y + 0.28*s + by // head center follows the bob

	p.fill(x+0.36*s, y+0.74*s, 0.09*s, 0.18*s, shade(ink, 0.5)) // planted legs
	p.fill(x+0.55*s, y+0.74*s, 0.09*s, 0.18*s, shade(ink, 0.5))
	p.fill(x+0.3*s, y+0.44*s+by*0.5, 0.4*s, 0.32*s, ink) // torso
	p.disc(cx, hy, 0.16*s, skin)
	if !blinkClosed(t, phase) {
		p.fill(cx-0.09*s, hy-0.02*s, 0.045*s, 0.045*s, darkEye)
		p.fill(cx+0.045*s, hy-0.02*s, 0.045*s, 0.045*s, darkEye)
	}

	switch class {
	case "Fighter":
		p.fill(x+0.32*s, y+0.13*s+by, 0.36*s, 0.12*s, steel) // helmet
		p.fill(x+0.34*s, y+0.18*s+by, 0.32*s, 0.035*s, darkEye)
		p.fill(x+0.78*s, y+0.16*s, 0.05*s, 0.5*s, steel) // upright sword
		p.fill(x+0.72*s, y+0.64*s, 0.17*s, 0.045*s, gold)
		p.fill(x+0.785*s, y+0.685*s, 0.04*s, 0.1*s, shade(gold, 0.6))
		if k, act := sweep01(t, phase, 2.4); act { // glint runs up the blade
			p.glow(x+0.805*s, y+0.62*s-float32(k)*0.42*s, 0.05*s, white, 0.85)
		}
	case "Cleric":
		p.poly([][2]float32{{x + 0.28*s, y + 0.92*s}, {x + 0.72*s, y + 0.92*s}, {x + 0.64*s, y + 0.46*s + by}, {x + 0.36*s, y + 0.46*s + by}}, ink)
		p.fill(x+0.46*s, y+0.48*s+by*0.5, 0.08*s, 0.28*s, gold) // stole
		for i := 0; i < 3; i++ {                                // halo of light dots orbiting the head
			a := t*2 + phase + float64(i)*2.094
			p.glow(cx+0.18*s*float32(math.Cos(a)), y+0.04*s+by+0.05*s*float32(math.Sin(a)), 0.04*s, gold, 0.6+0.35*pulse01(t, phase+float64(i), 0.5))
		}
	case "Wizard":
		sway := 0.04 * s * float32(math.Sin(t*2+phase))
		p.poly([][2]float32{{x + 0.26*s, y + 0.92*s}, {x + 0.74*s, y + 0.92*s}, {x + 0.66*s, y + 0.46*s + by}, {x + 0.34*s, y + 0.46*s + by}}, ink)
		poly := [][2]float32{{x + 0.28*s, hy - 0.08*s}, {x + 0.72*s, hy - 0.08*s}, {cx + 0.06*s + sway, y - 0.06*s + by}} // pointed hat
		p.poly(poly, ink)
		p.fill(x+0.26*s, hy-0.12*s, 0.48*s, 0.05*s, shade(ink, 1.3)) // brim
		p.glow(cx+0.06*s+sway, y-0.05*s+by, 0.045*s, white, 0.4+0.6*pulse01(t, phase, 1.2))
		p.fill(x+0.8*s, y+0.3*s, 0.035*s, 0.6*s, rgb{0.45, 0.32, 0.2}) // staff
		p.glow(x+0.817*s, y+0.26*s, 0.06*s, rgb{0.6, 0.75, 1}, 0.5+0.5*pulse01(t, phase+1, 0.9))
	case "Rogue":
		p.poly([][2]float32{{x + 0.24*s, hy + 0.06*s}, {x + 0.76*s, hy + 0.06*s}, {cx, y + 0.04*s + by}}, ink) // hood
		p.fill(x+0.3*s, hy+0.04*s, 0.4*s, 0.06*s, shade(ink, 0.7))                                             // scarf hides the face
		if !blinkClosed(t, phase) {
			p.fill(cx+0.045*s, hy-0.02*s, 0.045*s, 0.045*s, darkEye) // one watchful eye
		}
		p.poly([][2]float32{{x + 0.8*s, y + 0.36*s}, {x + 0.84*s, y + 0.4*s}, {x + 0.74*s, y + 0.62*s}, {x + 0.7*s, y + 0.58*s}}, steel) // dagger
		if k, act := sweep01(t, phase+0.5, 3.3); act {                                                                                   // brief edge flash
			p.glow(x+0.77*s, y+(0.42-0.12*float32(k))*s, 0.04*s, white, 0.9)
		}
	}
}

// drawEnemyAvatar paints each bestiary species with a body plan and an
// idle of its own, over the same base hue the flat boxes used.
func drawEnemyAvatar(c *DungeonCrawler, kind string, x, y, s float32, phase, t float64, dim bool) {
	p := pen{c, dim}
	base := enemyColor(kind)
	by := bob(t, phase)

	switch kind {
	case "rat":
		twitch := float32(math.Sin(t*5 + phase))
		p.fill(x+0.2*s, y+0.56*s, 0.46*s, 0.2*s, base) // low body
		p.disc(x+0.66*s, y+0.62*s, 0.12*s, base)       // snout end
		p.disc(x+0.56*s, y+0.46*s, 0.06*s, shade(base, 1.3))
		p.disc(x+0.7*s, y+0.44*s, 0.06*s, shade(base, 1.3)) // round ears
		p.disc(x+0.74*s, y+0.6*s, 0.025*s, rgb{0.9, 0.25, 0.25})
		p.fill(x+0.06*s, y+0.62*s+0.05*s*twitch, 0.16*s, 0.035*s, shade(base, 0.75)) // flicking tail
		p.fill(x+0.0*s, y+0.66*s-0.06*s*twitch, 0.1*s, 0.035*s, shade(base, 0.75))
	case "goblin":
		twitch := float32(0)
		if math.Sin(t*7+phase) > 0.72 { // ear twitch bursts
			twitch = 0.06 * s
		}
		p.poly([][2]float32{{x + 0.28*s, y + 0.26*s + by}, {x + 0.12*s, y + 0.1*s + by - twitch}, {x + 0.34*s, y + 0.18*s + by}}, base)
		p.poly([][2]float32{{x + 0.72*s, y + 0.26*s + by}, {x + 0.88*s, y + 0.1*s + by - twitch}, {x + 0.66*s, y + 0.18*s + by}}, base) // big ears
		p.fill(x+0.34*s, y+0.5*s+by*0.5, 0.32*s, 0.3*s, shade(base, 0.8))
		p.disc(x+0.5*s, y+0.32*s+by, 0.16*s, base)
		p.fill(x+0.42*s, y+0.3*s+by, 0.05*s, 0.05*s, rgb{0.9, 0.3, 0.2})
		p.fill(x+0.54*s, y+0.3*s+by, 0.05*s, 0.05*s, rgb{0.9, 0.3, 0.2})
		p.fill(x+0.36*s, y+0.78*s, 0.28*s, 0.1*s, shade(base, 0.55))
	case "skeleton":
		rattle := 0.7 * float32(math.Sin(t*30+phase)) // faint bone rattle
		p.disc(x+0.5*s, y+0.28*s+by, 0.15*s, base)
		p.fill(x+0.44*s, y+0.26*s+by, 0.04*s, 0.05*s, darkEye) // sockets stay dark
		p.fill(x+0.54*s, y+0.26*s+by, 0.04*s, 0.05*s, darkEye)
		p.fill(x+0.48*s, y+0.42*s+by*0.5, 0.04*s, 0.36*s, base) // spine
		for i := 0; i < 3; i++ {                                // ribs
			p.fill(x+0.36*s+float32(i)*rattle*0.3, y+(0.48+0.09*float32(i))*s+by*0.5, 0.28*s, 0.03*s, shade(base, 0.9))
		}
		p.fill(x+0.4*s, y+0.78*s, 0.2*s, 0.06*s, base)
	case "orc":
		bre := float32(pulse01(t, phase, 0.5)-0.5) * 0.07 * s // heavy breathing
		p.fill(x+0.22*s, y+0.44*s+by*0.5-bre, 0.56*s, 0.38*s+bre, base)
		p.disc(x+0.4*s, y+0.34*s+by, 0.1*s, shade(base, 0.7)) // shoulder pads
		p.disc(x+0.62*s, y+0.34*s+by, 0.1*s, shade(base, 0.7))
		p.disc(x+0.5*s, y+0.26*s+by, 0.13*s, shade(base, 1.15))
		p.fill(x+0.45*s, y+0.24*s+by, 0.04*s, 0.04*s, rgb{0.95, 0.85, 0.2})
		p.fill(x+0.53*s, y+0.24*s+by, 0.04*s, 0.04*s, rgb{0.95, 0.85, 0.2})
		p.fill(x+0.46*s, y+0.34*s+by, 0.03*s, 0.05*s, white) // tusks
		p.fill(x+0.52*s, y+0.34*s+by, 0.03*s, 0.05*s, white)
	case "shadow":
		fl := float32(pulse01(t, phase, 0.7))
		fy := y + 0.44*s - 2.5*fl // floats instead of bobbing
		p.glow(x+0.5*s, fy, 0.3*s, base, 0.45+0.25*float64(fl))
		poly := [][2]float32{{x + 0.5*s, fy - 0.28*s}, {x + 0.78*s, fy}, {x + 0.5*s, fy + 0.3*s}, {x + 0.22*s, fy}}
		p.glow(x+0.5*s, fy+0.16*s, 0.16*s+0.05*s*float32(math.Sin(t*2+phase)), base, 0.35) // wisping tail
		p.polyGlow(poly, shade(base, 1.2), 0.65+0.2*float64(fl))
		p.glow(x+0.44*s, fy-0.04*s, 0.035*s, rgb{0.75, 0.6, 1}, 0.9)
		p.glow(x+0.58*s, fy-0.04*s, 0.035*s, rgb{0.75, 0.6, 1}, 0.9)
	case "troll":
		bre := float32(pulse01(t, phase, 0.35)-0.5) * 0.09 * s
		p.fill(x+0.16*s, y+0.36*s+by*0.4-bre, 0.68*s, 0.5*s+bre, base) // hulking torso
		p.fill(x+0.06*s, y+0.44*s, 0.12*s, 0.34*s, shade(base, 0.8))   // thick arms
		p.fill(x+0.82*s, y+0.44*s, 0.12*s, 0.34*s, shade(base, 0.8))
		p.disc(x+0.5*s, y+0.28*s+by, 0.14*s, shade(base, 1.1))
		p.fill(x+0.4*s, y+0.2*s+by, 0.2*s, 0.03*s, shade(base, 0.5)) // heavy brow
		p.fill(x+0.44*s, y+0.26*s+by, 0.045*s, 0.04*s, rgb{0.95, 0.6, 0.15})
		p.fill(x+0.54*s, y+0.26*s+by, 0.045*s, 0.04*s, rgb{0.95, 0.6, 0.15})
	case "hag":
		p.disc(x+0.56*s, y+0.62*s+by*0.4, 0.24*s, base) // hunched back
		p.disc(x+0.4*s, y+0.38*s+by, 0.14*s, shade(base, 1.15))
		p.poly([][2]float32{{x + 0.26*s, y + 0.3*s + by}, {x + 0.54*s, y + 0.3*s + by}, {x + 0.66*s, y + 0.06*s + by}}, shade(base, 0.7)) // crooked hat
		p.fill(x+0.34*s, y+0.36*s+by, 0.05*s, 0.04*s, darkEye)
		p.fill(x+0.3*s, y+0.44*s+by, 0.09*s, 0.03*s, shade(base, 0.5)) // hooked nose
		gl := 0.4 + 0.5*pulse01(t, phase, 0.8)
		p.glow(x+0.2*s, y+0.72*s, 0.09*s, rgb{0.5, 0.95, 0.5}, gl) // bubbling orb
		p.disc(x+0.2*s, y+0.72*s, 0.05*s, rgb{0.7, 1, 0.7})
	case "wraith":
		fl := float32(pulse01(t, phase, 0.45))
		fy := by - 3*fl // drifts above its tile
		poly := [][2]float32{{x + 0.42*s, y + 0.2*s + fy}, {x + 0.58*s, y + 0.2*s + fy}, {x + 0.7*s, y + 0.9*s + fy}, {x + 0.3*s, y + 0.9*s + fy}}
		p.poly(poly, base)
		for i := 0; i < 3; i++ { // torn hem fluttering
			tag := 0.06*s + 0.05*s*float32(math.Sin(t*3+phase+float64(i)))
			p.poly([][2]float32{{x + (0.34+0.16*float32(i))*s, y + 0.86*s + fy}, {x + (0.44+0.16*float32(i))*s, y + 0.86*s + fy}, {x + (0.39+0.16*float32(i))*s, y + 0.86*s + fy + tag}}, shade(base, 0.85))
		}
		p.disc(x+0.5*s, y+0.3*s+fy, 0.12*s, rgb{0.06, 0.08, 0.12}) // hollow face
		p.glow(x+0.46*s, y+0.29*s+fy, 0.03*s, rgb{0.8, 0.95, 1}, 0.95)
		p.glow(x+0.55*s, y+0.29*s+fy, 0.03*s, rgb{0.8, 0.95, 1}, 0.95)
		p.glow(x+0.5*s, y+0.96*s+fy, 0.07*s, base, 0.3) // mist trailing off
	default: // the Grim Warden: aura, crown, and burning eyes
		au := pulse01(t, phase, 0.8)
		p.glow(x+0.5*s, y+0.52*s, 0.55*s, base, 0.18+0.15*au)
		p.fill(x+0.18*s, y+0.34*s+by*0.3, 0.64*s, 0.56*s, shade(base, 0.45)) // iron frame
		p.disc(x+0.2*s, y+0.38*s+by*0.3, 0.12*s, shade(base, 0.6))
		p.disc(x+0.8*s, y+0.38*s+by*0.3, 0.12*s, shade(base, 0.6)) // pauldrons
		p.disc(x+0.5*s, y+0.3*s+by, 0.16*s, shade(base, 0.55))
		for i := 0; i < 4; i++ { // crown spikes
			sx := x + (0.34+0.11*float32(i))*s
			p.poly([][2]float32{{sx, y + 0.18*s + by}, {sx + 0.06*s, y + 0.18*s + by}, {sx + 0.03*s, y + 0.04*s + by}}, shade(base, 0.7))
		}
		p.glow(x+0.45*s, y+0.3*s+by, 0.045*s, rgb{1, 0.2, 0.15}, 0.7+0.3*au)
		p.glow(x+0.57*s, y+0.3*s+by, 0.045*s, rgb{1, 0.2, 0.15}, 0.7+0.3*au)
	}
}

// drawItemAvatar gives floor features a figure: sloshing potion, shining
// weapon, shimmering stairs, straining door, chomping trap.
func drawItemAvatar(c *DungeonCrawler, f crawler.Feature, x, y, s float32, phase, t float64, dim bool) {
	p := pen{c, dim}
	switch f {
	case crawler.FeaturePotion:
		p.fill(x+0.44*s, y+0.32*s, 0.12*s, 0.08*s, rgb{0.5, 0.35, 0.22})                                    // cork
		p.fill(x+0.45*s, y+0.4*s, 0.1*s, 0.14*s, rgb{0.8, 0.85, 0.9})                                       // neck
		p.disc(x+0.5*s, y+0.64*s, 0.2*s, rgb{0.8, 0.85, 0.9})                                               // glass
		p.disc(x+0.5*s, y+0.68*s, 0.15*s, rgb{0.85, 0.35, 0.6})                                             // liquid
		p.fill(x+0.38*s, y+0.58*s+0.02*s*float32(math.Sin(t*3+phase)), 0.24*s, 0.03*s, rgb{0.92, 0.5, 0.7}) // sloshing surface
		p.glow(x+0.6*s, y+0.5*s, 0.03*s, white, 0.3+0.7*pulse01(t, phase, 1.1))                             // sparkle
		p.glow(x+0.4*s, y+0.72*s, 0.025*s, white, 0.3+0.7*pulse01(t, phase+2, 1.4))
	case crawler.FeatureWeapon:
		p.poly([][2]float32{{x + 0.5*s, y + 0.1*s}, {x + 0.58*s, y + 0.22*s}, {x + 0.58*s, y + 0.64*s}, {x + 0.42*s, y + 0.64*s}, {x + 0.42*s, y + 0.22*s}}, steel) // blade
		p.fill(x+0.32*s, y+0.64*s, 0.36*s, 0.06*s, gold)                                                                                                            // crossguard
		p.fill(x+0.46*s, y+0.7*s, 0.08*s, 0.16*s, rgb{0.45, 0.3, 0.18})                                                                                             // grip
		p.disc(x+0.5*s, y+0.88*s, 0.04*s, gold)
		if k, act := sweep01(t, phase+0.2, 2.2); act { // shine sweeping down the edge
			p.glow(x+0.5*s, y+float32(0.16+0.5*k)*s, 0.055*s, white, 0.8)
		}
	case crawler.FeatureStairs:
		p.fill(x+0.1*s, y+0.1*s, 0.8*s, 0.8*s, rgb{0.05, 0.07, 0.1}) // stairwell
		lit := int(t*2+phase) % 3
		for i := 0; i < 3; i++ {
			step := rgb{0.16, 0.5, 0.55}
			if i == lit {
				step = rgb{0.35, 0.85, 0.9} // shimmer walks down the steps
			}
			p.fill(x+0.16*s, y+(0.2+0.22*float32(i))*s, 0.68*s, 0.09*s, step)
		}
	case crawler.FeatureLockedDoor:
		shake := float32(0)
		if pulse01(t, phase, 0.35) > 0.93 { // occasionally strains against the lock
			shake = 0.6 * float32(math.Sin(t*24))
		}
		p.fill(x+0.12*s+shake*0.3, y+0.06*s, 0.76*s, 0.88*s, rgb{0.3, 0.2, 0.12}) // frame
		p.fill(x+0.2*s+shake, y+0.12*s, 0.26*s, 0.76*s, rgb{0.55, 0.38, 0.2})     // planks
		p.fill(x+0.54*s+shake, y+0.12*s, 0.26*s, 0.76*s, rgb{0.5, 0.34, 0.18})
		p.disc(x+0.64*s, y+0.5*s, 0.05*s, darkEye) // keyhole
		p.fill(x+0.625*s, y+0.5*s, 0.03*s, 0.09*s, darkEye)
	case crawler.FeatureOpenDoor:
		p.fill(x+0.12*s, y+0.06*s, 0.76*s, 0.88*s, rgb{0.1, 0.09, 0.12}) // open doorway
		sway := 0.4 * float32(math.Sin(t*1.2+phase))
		p.fill(x+0.16*s+sway, y+0.1*s, 0.18*s, 0.8*s, rgb{0.42, 0.32, 0.2}) // door leaf resting open
	case crawler.FeatureTrapArmed:
		gap := float32(0.14+0.12*math.Sin(t*6+phase)) * s // jaws chomp
		p.fill(x+0.18*s, y+0.5*s-gap/2-0.16*s, 0.64*s, 0.16*s, steel)
		p.fill(x+0.18*s, y+0.5*s+gap/2, 0.64*s, 0.16*s, steel)
		for i := 0; i < 3; i++ {
			tx := x + (0.26+0.2*float32(i))*s
			p.poly([][2]float32{{tx, y + 0.5*s - gap/2 - 0.16*s}, {tx + 0.08*s, y + 0.5*s - gap/2 - 0.16*s}, {tx + 0.04*s, y + 0.5*s - gap/2 - 0.02*s}}, shade(steel, 1.2))
			p.poly([][2]float32{{tx, y + 0.5*s + gap/2 + 0.16*s}, {tx + 0.08*s, y + 0.5*s + gap/2 + 0.16*s}, {tx + 0.04*s, y + 0.5*s + gap/2 + 0.02*s}}, shade(steel, 1.2))
		}
		if pulse01(t, phase, 0.9) > 0.9 {
			p.glow(x+0.5*s, y+0.5*s, 0.04*s, white, 0.8) // steel glint
		}
	case crawler.FeatureTrapDisabled:
		p.fill(x+0.18*s, y+0.2*s, 0.64*s, 0.14*s, shade(steel, 0.6)) // sprung and still
		p.fill(x+0.18*s, y+0.66*s, 0.64*s, 0.14*s, shade(steel, 0.6))
		p.fill(x+0.3*s, y+0.46*s, 0.4*s, 0.08*s, rgb{0.2, 0.2, 0.24})
	}
}
