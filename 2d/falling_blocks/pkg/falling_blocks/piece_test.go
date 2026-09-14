package fallingblocks

import "testing"

func TestKindName(t *testing.T) {
	want := map[Kind]string{
		KindI: "I",
		KindO: "O",
		KindT: "T",
		KindS: "S",
		KindZ: "Z",
		KindJ: "J",
		KindL: "L",
	}
	for k, w := range want {
		if got := KindName(k); got != w {
			t.Fatalf("KindName(%d) = %q, want %q", int(k), got, w)
		}
	}
}
