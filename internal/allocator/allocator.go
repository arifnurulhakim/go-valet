package allocator

import (
	"hash/fnv"
)

const (
	BasePort  = 9000
	PortRange = 1000
)

// GetPort returns a deterministic port based on the project path.
// Collision is possible but minimal with 1000 ports.
func GetPort(path string) int {
	h := fnv.New32a()
	h.Write([]byte(path))
	hash := h.Sum32()
	return BasePort + int(hash%PortRange)
}
