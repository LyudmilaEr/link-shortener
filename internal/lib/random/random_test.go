package random

import "testing"

func TestNewRandom(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "size = 0", size: 0},
		{name: "size = 1", size: 1},
		{name: "size = 5", size: 5},
		{name: "size = 10", size: 10},
		{name: "size = 20", size: 20},
		{name: "size = 30", size: 30},
	}

	allowed := map[rune]struct{}{}
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789" {
		allowed[r] = struct{}{}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewRandomString(tt.size)
			if len(s) != tt.size {
				t.Fatalf("unexpected length: got %d, want %d", len(s), tt.size)
			}

			for _, r := range s {
				if _, ok := allowed[r]; !ok {
					t.Fatalf("string contains invalid rune %q", r)
				}
			}
		})
	}
}
