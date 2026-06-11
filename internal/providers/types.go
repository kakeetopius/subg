// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"errors"

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
	providers []subtitleProvider

	options Options
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

func NewProviderSet() ProviderSet {
	return ProviderSet{}
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

func (set *ProviderSet) Start() error {
	for _, provider := range set.providers {
		provider.WithOptions(set.options)

		if provider.providerSpecificOpts != nil {
			provider.WithSpecificOptions(provider.providerSpecificOpts)
		}

		err := provider.SearchSubtitle()
		if err != nil {
			pterm.Error.Printf("Provider %s returned error: %s\n", provider.Name(), err)
			continue
		}

		if set.options.AutoSelect {
			err = provider.DownloadBest()
			if err != nil {
				pterm.Error.Printf("Provider %s returned error: %s\n", provider.Name(), err)
				continue
			}
			return nil
		}

		selected, err := provider.DisplaySelections()
		if err != nil {
			if !errors.Is(err, ErrNextProvider) {
				pterm.Error.Printf("Provider %s returned error: %s\n", provider.Name(), err)
			}
			continue
		}

		err = provider.Download(selected)
		if err != nil {
			pterm.Error.Printf("Provider %s returned error: %s\n", provider.Name(), err)
			continue
		}
		return nil
	}

	return nil
}
