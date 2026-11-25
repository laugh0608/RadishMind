//go:build darwin

package developmentcontrol

import "golang.org/x/sys/unix"

func processNamespace() string {
	boot, err := unix.Sysctl("kern.bootsessionuuid")
	if err != nil {
		return ""
	}
	return boot
}
