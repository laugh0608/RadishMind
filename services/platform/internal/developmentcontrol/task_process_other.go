//go:build !linux && !darwin

package developmentcontrol

func processNamespace() string { return "" }
func processAbsent(int) bool   { return false }
