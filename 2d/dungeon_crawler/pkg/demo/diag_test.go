package demo

import (
	"testing"

	"godot-go-demo-projects/2d/dungeon_crawler/pkg/crawler"
)

// --- 19.2 Shift-diagonal input resolution ---

// Shift tilts a direction one step clockwise: ↑↗ →↘ ↓↙ ←↖.
func TestDiagonalForRotatesClockwise(t *testing.T) {
	cases := []struct {
		pressed crawler.Dir
		want    crawler.Dir
	}{
		{crawler.Up, crawler.UpRight},
		{crawler.Right, crawler.DownRight},
		{crawler.Down, crawler.DownLeft},
		{crawler.Left, crawler.UpLeft},
	}
	for _, c := range cases {
		if got := diagonalFor(c.pressed); got != c.want {
			t.Errorf("diagonalFor(%v) = %v, want %v", c.pressed, got, c.want)
		}
	}
}
