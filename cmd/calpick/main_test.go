package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseArguments(t *testing.T) {
	nov15 := time.Date(2026, time.November, 15, 0, 0, 0, 0, time.Local)
	nov20 := time.Date(2026, time.November, 20, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name    string
		args    []string
		want    cliArgs
		wantErr string
	}{
		{
			name: "no arguments",
			args: nil,
		},
		{
			name: "single date",
			args: []string{"2026-11-15"},
			want: cliArgs{from: nov15, hasFrom: true},
		},
		{
			name:    "too many single-mode dates",
			args:    []string{"2026-11-15", "2026-11-20"},
			wantErr: "too many arguments",
		},
		{
			name: "range without dates",
			args: []string{"--range"},
			want: cliArgs{rangeMode: true},
		},
		{
			name: "version",
			args: []string{"--version"},
			want: cliArgs{version: true},
		},
		{
			name: "short version",
			args: []string{"-v"},
			want: cliArgs{version: true},
		},
		{
			name: "version with extra arguments",
			args: []string{"--range", "--version", "extra"},
			want: cliArgs{rangeMode: true, version: true},
		},
		{
			name: "help wins over version",
			args: []string{"--version", "--help"},
			want: cliArgs{version: true, help: true},
		},
		{
			name: "range with two dates",
			args: []string{"--range", "2026-11-15", "2026-11-20"},
			want: cliArgs{rangeMode: true, from: nov15, until: nov20, hasFrom: true},
		},
		{
			name: "range with reversed dates is parsed as given",
			args: []string{"--range", "2026-11-20", "2026-11-15"},
			want: cliArgs{rangeMode: true, from: nov20, until: nov15, hasFrom: true},
		},
		{
			name:    "range with one date",
			args:    []string{"--range", "2026-11-15"},
			wantErr: "zero or two dates",
		},
		{
			name: "flag after dates",
			args: []string{"2026-11-15", "2026-11-20", "--range"},
			want: cliArgs{rangeMode: true, from: nov15, until: nov20, hasFrom: true},
		},
		{
			name: "help",
			args: []string{"--help"},
			want: cliArgs{help: true},
		},
		{
			name: "help with extra arguments",
			args: []string{"--range", "-h", "extra"},
			want: cliArgs{rangeMode: true, help: true},
		},
		{
			name:    "unknown option",
			args:    []string{"--rnage", "2026-11-15"},
			wantErr: "unknown option",
		},
		{
			name:    "invalid range end date",
			args:    []string{"--range", "2026-11-15", "2026-13-01"},
			wantErr: "invalid date",
		},
		{
			name:    "invalid date",
			args:    []string{"2026-13-01"},
			wantErr: "invalid date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArguments(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.rangeMode != tt.want.rangeMode || got.help != tt.want.help || got.version != tt.want.version || got.hasFrom != tt.want.hasFrom {
				t.Fatalf("args = %+v, want %+v", got, tt.want)
			}
			if !got.from.Equal(tt.want.from) || !got.until.Equal(tt.want.until) {
				t.Fatalf("dates = %v..%v, want %v..%v", got.from, got.until, tt.want.from, tt.want.until)
			}
		})
	}
}

func TestVersionLine(t *testing.T) {
	if got, want := versionLine(), "calpick v0.1.1"; got != want {
		t.Fatalf("versionLine() = %q, want %q", got, want)
	}
}
