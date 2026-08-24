package validation

import "testing"

func TestMatchRegex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
		want    bool
		wantErr bool
	}{
		{"match", "^ab.*", "abcdef", true, false},
		{"no match", "^zz", "abcdef", false, false},
		{"digits", `\d{3}`, "id=404", true, false},
		{"invalid pattern", "(", "abc", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchRegex(tt.pattern, tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("matchRegex err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("matchRegex(%q,%q) = %v, want %v", tt.pattern, tt.text, got, tt.want)
			}
		})
	}
}
