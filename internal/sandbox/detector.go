package sandbox

import (
	"bytes"
	"os/exec"
)

// Detect returns true if the Apple Container system is installed and running.
// It checks `container system status` for "running" in the output.
func Detect() bool {
	path, err := exec.LookPath("container")
	if err != nil {
		return false
	}
	out, err := exec.Command(path, "system", "status").Output()
	if err != nil {
		return false
	}
	return bytes.Contains(out, []byte("running"))
}
