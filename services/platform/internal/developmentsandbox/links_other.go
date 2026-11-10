//go:build !linux && !darwin

package developmentsandbox

import "os"

func singleLink(os.FileInfo) bool { return false }

func ownedFile(os.FileInfo) bool { return false }
