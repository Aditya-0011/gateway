package db

import (
	"strings"
	"testing"
)

func TestParseCookie(t *testing.T) {
	token := strings.Repeat("a", 43)
	tests := []struct {
		name   string
		value  string
		wantID int
		wantOK bool
	}{
		{"valid", "42." + token, 42, true},
		{"empty", "", 0, false},
		{"no separator", "42" + token, 0, false},
		{"non numeric id", "abc." + token, 0, false},
		{"zero id", "0." + token, 0, false},
		{"negative id", "-1." + token, 0, false},
		{"short token", "42." + strings.Repeat("a", 31), 0, false},
		{"empty id", "." + token, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, sid, ok := ParseCookie(tt.value)
			if ok != tt.wantOK || id != tt.wantID {
				t.Fatalf("ParseCookie(%q) = %d, %q, %v; want %d, ok=%v", tt.value, id, sid, ok, tt.wantID, tt.wantOK)
			}
			if ok && sid != sessionID(token) {
				t.Fatalf("sid = %q, want hash of token", sid)
			}
		})
	}
}
