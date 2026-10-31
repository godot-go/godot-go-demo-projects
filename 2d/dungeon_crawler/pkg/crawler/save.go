package crawler

import (
	"encoding/json"
	"os"
)

// Meta is the game's only persisted state: renown, purchased building levels,
// ever-killed specials, and the selected class. The shell stores it as JSON at
// a user:// path it owns; the rules layer only sees the path string.
type Meta struct {
	Renown     int
	Buildings  map[string]int
	FirstKills map[string]bool
	Selected   string
	Flavor     *CharacterFlavor `json:",omitempty"`
}

// DefaultMeta is the fresh-install state: nothing earned, every building at
// level 0, and only the starter class unlocked (and pre-selected).
func DefaultMeta() *Meta {
	m := &Meta{
		Buildings:  map[string]int{},
		FirstKills: map[string]bool{},
	}
	for name := range upgrades {
		m.Buildings[name] = 0
	}
	m.Selected = starterClass
	return m
}

// LoadMeta reads the JSON save. A missing file yields fresh-install defaults
// rather than an error so first launch is uneventful.
func LoadMeta(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultMeta(), nil
		}
		return nil, err
	}
	m := DefaultMeta()
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	if m.Buildings == nil {
		m.Buildings = map[string]int{}
	}
	if m.FirstKills == nil {
		m.FirstKills = map[string]bool{}
	}
	if m.Selected == "" {
		m.Selected = starterClass
	}
	return m, nil
}

// SaveMeta writes the save file.
func SaveMeta(path string, m *Meta) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// RecordFirstKill marks a special kill; it reports whether this one was the
// first ever (which pays the renown bonus).
func (m *Meta) RecordFirstKill(kind string) bool {
	if m.FirstKills[kind] {
		return false
	}
	m.FirstKills[kind] = true
	return true
}

// IsFresh reports whether the meta carries no progress worth warning about
// when a new game would overwrite it: no renown, no purchased levels, no
// first kills. A bare starter-class selection alone never counts as played.
func (m *Meta) IsFresh() bool {
	if m.Renown != 0 {
		return false
	}
	for _, lvl := range m.Buildings {
		if lvl != 0 {
			return false
		}
	}
	return len(m.FirstKills) == 0
}
