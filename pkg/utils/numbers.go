package utils

// IntOrEmptyInt returns an empty int if a int pointer is nil.
// Otherwise returns the int value of the pointer.
func IntOrEmptyInt(s *int) int {
	if s != nil {
		return *s
	}
	return 0
}

// IntToPtr is a convenience func that converts a int to it's pointer
func IntToPtr(s int) *int {
	return &s
}
