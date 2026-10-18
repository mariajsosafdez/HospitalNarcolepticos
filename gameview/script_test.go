package gameview

import (
	"strings"
	"testing"
)

func TestParseScript(t *testing.T) {
	src := `# comentario
wait 120; click 100 50
key Enter
shot out/a.png
shot4 out/b.png ; dump out/c.json
sleep P-001 2
wake P-001
move 5 6
`
	steps, err := parseScript(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []step{
		{kind: stepWait, n: 120, line: 2},
		{kind: stepClick, x: 100, y: 50, line: 2},
		{kind: stepKey, text: "Enter", line: 3},
		{kind: stepShot, text: "out/a.png", line: 4},
		{kind: stepShot4, text: "out/b.png", line: 5},
		{kind: stepDump, text: "out/c.json", line: 5},
		{kind: stepSleep, text: "P-001", cam: "2", line: 6},
		{kind: stepWake, text: "P-001", line: 7},
		{kind: stepMove, x: 5, y: 6, line: 8},
	}
	if len(steps) != len(want) {
		t.Fatalf("got %d steps, want %d: %+v", len(steps), len(want), steps)
	}
	for i := range want {
		if steps[i] != want[i] {
			t.Errorf("step %d = %+v, want %+v", i, steps[i], want[i])
		}
	}
}

func TestParseScriptErrors(t *testing.T) {
	tests := []struct{ src, contains string }{
		{"dance 3", `unknown step "dance"`},
		{"key Banana", `unknown key "Banana"`},
		{"wait", "needs 1 argument"},
		{"wait x", "not a number"},
		{"wait -3", "negative"},
		{"click 4", "needs 2 argument"},
		{"shot", "needs 1 argument"},
		{"sleep P-001", "needs 2 argument"},
	}
	for _, tt := range tests {
		_, err := parseScript(tt.src)
		if err == nil || !strings.Contains(err.Error(), tt.contains) {
			t.Errorf("parseScript(%q) error = %v, want it to contain %q", tt.src, err, tt.contains)
		}
	}
}
