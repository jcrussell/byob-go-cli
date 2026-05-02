// Package site implements `byob site`: render decisions/ + memories/ as a
// browsable static site under -out (default _site).
package site

import (
	"github.com/spf13/cobra"

	sitepkg "github.com/jcrussell/byob-go-cli/internal/site"
	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

type Options struct {
	IO           *iostreams.IOStreams
	DecisionsDir string
	MemoriesDir  string
	ReadmePath   string
	OutDir       string
	BaseURL      string
	RepoURL      string
	Strict       bool
}

func NewCmdSite(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{IO: f.IOStreams}
	cmd := &cobra.Command{
		Use:   "site",
		Short: "Render decisions/ + memories/ as a static site",
		Long: "Walks decisions/<cat>/*.md and memories/*.md, rewrites bare " +
			"`byob-foo.N` cross-references to anchor links, renders markdown " +
			"with goldmark + Chroma, and writes one index.html per route.",
		RunE: func(c *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return siteRun(opts)
		},
	}
	cmd.Flags().StringVar(&opts.DecisionsDir, "decisions-dir", "decisions", "decisions tree root")
	cmd.Flags().StringVar(&opts.MemoriesDir, "memories-dir", "memories", "memories directory")
	cmd.Flags().StringVar(&opts.ReadmePath, "readme", "README.md", "README to source the homepage intro from (empty to skip)")
	cmd.Flags().StringVar(&opts.OutDir, "out", "_site", "output directory")
	cmd.Flags().StringVar(&opts.BaseURL, "base-url", "", "URL prefix for absolute links (e.g. /byob-go-cli for project pages)")
	cmd.Flags().StringVar(&opts.RepoURL, "repo-url", "https://github.com/jcrussell/byob-go-cli", "GitHub repo URL for header and source links")
	cmd.Flags().BoolVar(&opts.Strict, "strict", false, "fail the build on unknown byob-* cross-references")
	return cmd
}

func siteRun(opts *Options) error {
	s, err := sitepkg.Walk(opts.DecisionsDir, opts.MemoriesDir)
	if err != nil {
		return err
	}
	s.BaseURL = opts.BaseURL
	s.RepoURL = opts.RepoURL
	return sitepkg.Render(s, opts.OutDir, opts.ReadmePath, opts.Strict, opts.IO.ErrOut)
}
