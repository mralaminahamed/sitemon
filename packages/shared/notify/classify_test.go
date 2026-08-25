package notify

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		prev, status, want string
	}{
		{"", "UP", ""},   // first sighting, up
		{"", "DOWN", ""}, // first sighting, down
		{"UP", "UP", ""}, // steady up
		{"UP", "DOWN", "down"},
		{"DOWN", "DOWN", ""}, // steady down
		{"DOWN", "UP", "recovery"},
		{"UP", "WARNING", "down"},
		{"WARNING", "UP", "recovery"},
	}
	for _, c := range cases {
		if got := Classify(c.prev, c.status); got != c.want {
			t.Errorf("Classify(%q,%q)=%q want %q", c.prev, c.status, got, c.want)
		}
	}
}
