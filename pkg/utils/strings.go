package utils

// StrOrEmptyStr returns an empty string if a string pointer is nil.
// Otherwise returns the string value of the pointer.
func StrOrEmptyStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// StrToPtr is a convenience func that converts a string to it's pointer
func StrToPtr(s string) *string {
	return &s
}
