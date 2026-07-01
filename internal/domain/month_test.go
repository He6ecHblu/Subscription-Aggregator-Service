package domain

import "testing"

func TestParseMonth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid", input: "07-2025", want: "07-2025"},
		{name: "single digit month", input: "7-2025", wantErr: true},
		{name: "invalid month", input: "13-2025", wantErr: true},
		{name: "wrong separator", input: "07/2025", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMonth(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestMonthMonthsUntilInclusive(t *testing.T) {
	t.Parallel()

	start, err := ParseMonth("07-2025")
	if err != nil {
		t.Fatalf("parse start month: %v", err)
	}

	end, err := ParseMonth("12-2025")
	if err != nil {
		t.Fatalf("parse end month: %v", err)
	}

	got, err := start.MonthsUntilInclusive(end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 6 {
		t.Fatalf("got %d, want 6", got)
	}
}
