package cmdutil

import "github.com/jcrussell/byob-go-cli/pkg/iostreams"

// Factory bundles cross-cutting dependencies. Built once in main(),
// threaded into every NewCmdXxx constructor. Eager fields are cheap;
// future expensive deps go in as `func() (T, error)` lazy closures.
// See byob-factory-di.1.
type Factory struct {
	IOStreams      *iostreams.IOStreams
	ExecutableName string
}

func New() *Factory {
	return &Factory{
		IOStreams:      iostreams.System(),
		ExecutableName: "byob",
	}
}
