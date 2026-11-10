// This binary is a container-only watchdog, never a host command execution API.
package main

import (
	"os"
	"radishmind.local/services/platform/internal/developmentsandboxrunner"
	"runtime"
	"strconv"
)

func main() {
	if runtime.GOOS != "linux" || len(os.Args) != 2 {
		os.Exit(125)
	}
	seconds, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(125)
	}
	os.Exit(developmentsandboxrunner.Run(seconds, os.Stdout))
}
