package pkg

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 3, "hel..."},
		{"world", 10, "world"},
		{"GoLang", 2, "Go..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Truncate(tt.input, tt.n)
			if got != tt.want {
				t.Errorf("Truncate(%s, %d) = %s; want %s", tt.input, tt.n, got, tt.want)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		s    string
		ss   []string
		want bool
	}{
		{"hello", []string{"hello", "world"}, true},
		{"go", []string{"golang", "python", "java"}, false},
		{"java", []string{"golang", "python", "java"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := StringInSlice(tt.s, tt.ss)
			if got != tt.want {
				t.Errorf("StringInSlice(%s, %v) = %v; want %v", tt.s, tt.ss, got, tt.want)
			}
		})
	}
}

func TestJudgeLogLevel(t *testing.T) {
	tests := []struct {
		line            string
		keywordPosition int
		want            string
	}{
		{"INFO: Everything is working fine.", 0, "info"},
		{"error: Failed to connect to the database.", 0, "error"},
		{"warn: Deprecated API usage.", 0, "warn"},
		{"fatal: Unexpected null pointer exception.", 0, "danger"},
		{"debug: Variable x has value 10.", 0, "debug"},
		{"INFO: Just another log entry.", 0, "info"},
		{"This is just a normal log entry.", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := JudgeLogLevel(tt.line, tt.keywordPosition)
			if got != tt.want {
				t.Errorf("JudgeLogLevel(%s, %d) = %s; want %s", tt.line, tt.keywordPosition, got, tt.want)
			}
		})
	}
}

func TestConsistentFormat(t *testing.T) {
	tests := []struct {
		logLines       []string
		wantConsistent bool
		wantPosition   int
	}{
		{
			[]string{
				"INFO: Everything is working fine.",
				"ERROR: Failed to connect to the database.",
				"WARNING: Deprecated API usage.",
				"FATAL: Unexpected null pointer exception.",
				"DEBUG: Variable x has value 10.",
			},
			true,
			0,
		},
		{
			[]string{
				"INFO - Everything is working fine.",
				"ERROR - Failed to connect to the database.",
				"WARNING - Deprecated API usage.",
				"FATAL - Unexpected null pointer exception.",
				"DEBUG - Variable x has value 10.",
			},
			true,
			0,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotConsistent, gotPosition := ConsistentFormat(tt.logLines)
			if gotConsistent != tt.wantConsistent || gotPosition != tt.wantPosition {
				t.Errorf("ConsistentFormat(%v) = %v, %d; want %v, %d", tt.logLines, gotConsistent, gotPosition, tt.wantConsistent, tt.wantPosition)
			}
		})
	}
}
