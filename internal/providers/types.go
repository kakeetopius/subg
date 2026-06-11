// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"errors"
	"fmt"

	"github.com/kakeetopius/subg/internal/ui"
	"github.com/pterm/pterm"
)

var ErrNextProvider = errors.New("user requested next provider")

type Provider interface {
	Name() string

	WithOptions(Options)

	WithSpecificOptions(any)

	SearchSubtitle() error

	DisplaySelections() ([]Subtitle, error)

	Download([]Subtitle) error

	DownloadBest() error
}

type Subtitle any

type ProviderSet struct {
	providers       []subtitleProvider
	options         Options
	subtitleQueries []string
}

type subtitleProvider struct {
	Provider
	providerSpecificOpts any
}

// Options are the general options for searching and downloading subtitles.
type Options struct {
	// Search Options
	Query    string
	IMDBId   int
	Season   int
	Episode  int
	Language string
	Type     string
	Year     int
	IsMovie  bool
	IsSerie  bool

	// Download Options
	SubtitleFormat string
	OutPutFile     string
	OutPutDir      string
	AutoSelect     bool
}

func NewProviderSet(subtitleQueries []string) ProviderSet {
	return ProviderSet{subtitleQueries: subtitleQueries}
}

func (set *ProviderSet) WithOptions(options Options) *ProviderSet {
	set.options = options
	return set
}

func (set *ProviderSet) WithProvider(provider Provider, providerSpecificOpts ...any) *ProviderSet {
	subProvider := subtitleProvider{
		Provider: provider,
	}
	if len(providerSpecificOpts) > 0 {
		subProvider.providerSpecificOpts = providerSpecificOpts[0]
	}

	set.providers = append(set.providers, subProvider)
	return set
}

func (set *ProviderSet) StartSearchAndDownload() error {
outer:
	for _, query := range set.subtitleQueries {
		fmt.Println()
		pterm.Info.Printf("Searching Subtitles for: %s\n", query)
		for _, provider := range set.providers {
			pterm.Info.Printf("Trying provider: %s\n", provider.Name())
			opts := set.options

			opts.Query = query
			err := searchAndDownloadWithProvider(provider, opts)
			if err != nil {
				if errors.Is(err, ui.ErrUserQuit) {
					return nil
				}
				if !errors.Is(err, ui.ErrNextProvider) {
					pterm.Error.Printf("Provider %s returned error: %s\n", provider.Name(), err)
				}
				continue
			}
			continue outer
		}
	}

	return nil
}

func searchAndDownloadWithProvider(provider subtitleProvider, opts Options) error {
	provider.WithOptions(opts)

	if provider.providerSpecificOpts != nil {
		provider.WithSpecificOptions(provider.providerSpecificOpts)
	}

	err := provider.SearchSubtitle()
	if err != nil {
		return err
	}

	if opts.AutoSelect {
		err = provider.DownloadBest()
		if err != nil {
			return err
		}
		return nil
	}

	selected, err := provider.DisplaySelections()
	if err != nil {
		return err
	}

	return provider.Download(selected)
}
