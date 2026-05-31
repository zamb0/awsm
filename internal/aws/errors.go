package aws

import "errors"

// Sentinel errors used across the aws package. Callers should use errors.Is to
// detect them rather than string-matching.
var (
	// ErrProfileNotFound is returned when a profile cannot be located in the
	// AWS config or credentials files.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrNoActiveProfile is returned when no profile is currently set as the
	// default/active one in ~/.aws/credentials.
	ErrNoActiveProfile = errors.New("no active profile")
)
