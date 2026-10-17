package port

import "testing"

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{name: "dash form", args: []string{"8000-9000"}, wantStart: 8000, wantEnd: 9000},
		{name: "split form", args: []string{"8000", "9000"}, wantStart: 8000, wantEnd: 9000},
		{name: "bad form", args: []string{"8000"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parsePortRange(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if start != tt.wantStart || end != tt.wantEnd {
				t.Fatalf("got %d-%d, want %d-%d", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
