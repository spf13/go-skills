package claims

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// go-release/SKILL.md "GODEBUG and GOEXPERIMENT: Deadlines, Not Escapes" and go/SKILL.md
// "time Channels Are Always Unbuffered (Go 1.27)": asynctimerchan was removed in Go 1.27,
// and a godebug entry pinning it to an old value fails the build.
func TestRemovedGODEBUGPinFailsTheBuild(t *testing.T) {
	gocmd, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	write := func(name, text string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.go", "package main\n\nfunc main() {}\n")
	for value, wantFail := range map[string]bool{"1": true, "0": false} {
		write("go.mod", "module pin\n\ngo 1.27\n\ngodebug asynctimerchan="+value+"\n")
		cmd := exec.Command(gocmd, "build", "-o", os.DevNull, ".")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if wantFail && (err == nil || !strings.Contains(string(out), `removed GODEBUG "asynctimerchan"`)) {
			t.Errorf("asynctimerchan=%s: want a removed-GODEBUG build failure, got err=%v\n%s", value, err, out)
		}
		if !wantFail && err != nil {
			t.Errorf("asynctimerchan=%s (the final default): build failed: %v\n%s", value, err, out)
		}
	}
}
