package cmdutil

import "github.com/jcrussell/byob-go-cli/pkg/iostreams"

// Factory bundles cross-cutting dependencies. Built once in main(),
// threaded into every NewCmdXxx constructor. Eager fields are cheap;
// future expensive deps go in as `func() (T, error)` lazy closures.
// See byob-factory-di.1.
//
// Prompter, Config, Store, and HTTPClient from the decision sketch are
// intentionally absent: split/join/site are file-IO build commands and
// don't need them. Add when the first command does.
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
