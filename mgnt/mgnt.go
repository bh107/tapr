package mgnt

import "hpt.space/tapr/store/tape"

// Client defines an administrative interface.
type Client interface {
	// Status returns a list of known volumes.
	Volumes() ([]tape.Volume, error)
}
