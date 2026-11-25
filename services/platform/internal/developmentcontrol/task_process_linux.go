//go:build linux

package developmentcontrol

import (
	"os"
	"strings"
)

// A PID can designate different processes on a different boot or in a new
// controller container. Absence is evidence only in the original PID namespace.
func processNamespace() string {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return ""
	}
	ns, err := os.Readlink("/proc/self/ns/pid")
	if err != nil || ns == "" {
		return ""
	}
	return strings.TrimSpace(string(boot)) + "/" + ns
}
