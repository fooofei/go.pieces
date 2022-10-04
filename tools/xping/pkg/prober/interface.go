package prober

import (
	"context"
)

// Prober defines prober's common methods
type Prober interface {
	Name() string

	Example() string
	// Ready is Before Probe, setup the probe context ready
	Ready(ctx context.Context, addr string) error
	// Probe is do the probe operation
	// The prob may contains several steps, only timing steps which we want
	// return values
	//     @string the custom message text for print
	//     @error indicates this probe success or fail
	Probe(ctx context.Context, addr string) (string, error)

	Close() error
}
