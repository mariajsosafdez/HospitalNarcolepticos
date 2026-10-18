package gameview

import "testing"

func TestNextPatient(t *testing.T) {
	ids := []string{"A", "B", "C"}
	tests := []struct {
		ids     []string
		current string
		dir     int
		want    string
	}{
		{ids, "", 1, "A"},
		{ids, "", -1, "C"},
		{ids, "A", 1, "B"},
		{ids, "C", 1, "A"},
		{ids, "A", -1, "C"},
		{ids, "Z", 1, "A"}, // elegido que ya no está a la vista
		{nil, "A", 1, ""},
	}
	for _, tt := range tests {
		if got := nextPatient(tt.ids, tt.current, tt.dir); got != tt.want {
			t.Errorf("nextPatient(%v,%q,%d) = %q, want %q", tt.ids, tt.current, tt.dir, got, tt.want)
		}
	}
}

func TestNextCursor(t *testing.T) {
	tests := []struct{ current, dir, n, want int }{
		{-1, 1, 3, 0},
		{-1, -1, 3, 2},
		{0, 1, 3, 1},
		{2, 1, 3, 0},
		{0, -1, 3, 2},
		{0, 1, 0, -1},
	}
	for _, tt := range tests {
		if got := nextCursor(tt.current, tt.dir, tt.n); got != tt.want {
			t.Errorf("nextCursor(%d,%d,%d) = %d, want %d", tt.current, tt.dir, tt.n, got, tt.want)
		}
	}
}
