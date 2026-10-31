package crawler

import (
	"fmt"
	"math/rand"
)

// CharacterFlavor is the rolled persona card from character generation:
// pure flavor with no mechanical stat effect, but stable for the life of the
// character so the sheet and the card never disagree. Rolled from one seed;
// the card stays what the player confirmed, not what a later reroll shows.
type CharacterFlavor struct {
	Hometown string `json:"hometown"`
	Boon     string `json:"boon"`
	Tales    []int  `json:"tales"` // four d20 fortunes, display only
	Sigil    int    `json:"sigil"` // index into flavorSigils
}

// flavorHometowns lists the little places a descender could come from.
var flavorHometowns = []string{
	"Cinderhollow, the ash-fall mining camp",
	"Brandervale, where the tollhouse never sleeps",
	"Duskmere, a lake town that keeps its lamps lit",
	"Old Saltmoor, salt-bleached and stubborn",
	"Hollowfen, built over somebody's cellar — several somebodies",
	"Grimspital, half monastery, half market",
	"Whist Ford, where the river hums at dusk",
	"Emberlees, decided by a fire nobody mentions",
}

// flavorBoons lists the one-line luck (curse?) the card rolls.
var flavorBoons = []string{
	"the last rat you spared followed you out of the dark — it still watches your pack",
	"you found a coin that is always your exact change, no matter the purse",
	"doors you lean on creak exactly once, which has saved your neck twice",
	"you can taste lies; they fizz, like bad water",
	"the innkeeper's dog follows you exactly one town, every time",
	"your boots never quite dry, and the damp always warns you of slime",
	"you once talked your way past a warden by explaining your taxes — twice",
	"the moon seems to over your shoulder when the torches gutter",
}

// flavorSigils colors the card; sigil words double as tint names.
var flavorSigils = []string{
	"Crimson", "Verdant", "Gilded", "Azure", "Violet", "Ivory",
}

// RollFlavor decides a persona card from a seed: same seed, same card.
func RollFlavor(seed int64) CharacterFlavor {
	rng := rand.New(rand.NewSource(seed))
	card := CharacterFlavor{
		Hometown: flavorHometowns[rng.Intn(len(flavorHometowns))],
		Boon:     flavorBoons[rng.Intn(len(flavorBoons))],
		Sigil:    rng.Intn(len(flavorSigils)),
	}
	for i := 0; i < 4; i++ {
		card.Tales = append(card.Tales, rng.Intn(20)+1)
	}
	return card
}

// Card renders the flavor card as sheet text: three tight lines so the
// character sheet never spills past its modal panel.
func (f *CharacterFlavor) Card() string {
	return fmt.Sprintf("✦ %s Sigil ✦ · Tales %d · %d · %d · %d\nFrom %s.\nBoon: %s.",
		flavorSigils[f.Sigil%len(flavorSigils)],
		f.Tales[0], f.Tales[1], f.Tales[2], f.Tales[3],
		f.Hometown, f.Boon)
}
