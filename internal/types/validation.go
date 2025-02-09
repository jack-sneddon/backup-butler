// internal/types/validation.go
package types

type ValidationLevel string

const (
	Quick    ValidationLevel = "quick"    // Size and modification time only
	Standard ValidationLevel = "standard" // Includes partial content hash
	Deep     ValidationLevel = "deep"     // Full content verification
)

func IsValidLevel(level string) bool {
	switch ValidationLevel(level) {
	case Quick, Standard, Deep:
		return true
	}
	return false
}
