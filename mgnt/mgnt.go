package mgnt

import "tapr.space/store/tape"

// Client defines an administrative interface.
type Client interface {
	// Status returns a list of known volumes.
	Volumes() ([]tape.Volume, error)
}
