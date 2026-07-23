// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/ui"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

var ErrNextProvider = errors.New("user requested next provider")

// Provider searches for and downloads subtitles from a subtitle service.
type Provider interface {
	// Name returns the provider's display name.
	Name() string

	// SearchSubtitle searches for subtitles matching the supplied options and
	// caches the results for subsequent operations.
	SearchSubtitle(ctx context.Context, opts Options) error

	// PromptSelection presents the cached search results to the user, prompts
	// them to select one or more subtitles, and returns the selected subtitles.
	PromptSelection() ([]Subtitle, error)

	// Download downloads the specified subtitle.
	Download(ctx context.Context, subtitle Subtitle) (SubtitleFile, error)

	// DownloadBest automatically selects the best subtitle from the cached
	// search results according to the provider's selection criteria and
	// downloads it.
	DownloadBest(ctx context.Context) (SubtitleFile, error)
}

// Subtitle represents a provider-specific subtitle search result.
type Subtitle interface {
	IsSub() // just a marker
}

// SubtitleFile represents the contents of a downloaded subtitle.
type SubtitleFile struct {
	// Name is the suggested filename of the subtitle. Can be empty
	Name string

	// Type identifies the subtitle format.
	Type subformat.FormatType

	// ReadCloser provides access to the subtitle contents.
	io.ReadCloser
}

// ProviderSet coordinates searches and downloads across one or more subtitle providers.
type ProviderSet struct {
	// providers is the set of registered subtitle providers.
	providers []Provider

	// options are the search options supplied for the current operation.
	options Options

	// subtitleQueries are the subtitles to search for
	subtitleQueries []string
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
	SubtitleFormat subformat.FormatType
	OutputFile     string
	OutputDir      string
	AutoSelect     bool
}

func NewProviderSet(subtitleQueries []string) *ProviderSet {
	return &ProviderSet{
		subtitleQueries: subtitleQueries,
		providers:       make([]Provider, 0, 5),
	}
}

func (set *ProviderSet) WithOptions(options Options) *ProviderSet {
	if options.Season != 0 || options.Episode != 0 {
		options.IsSerie = true
	}
	set.options = options
	return set
}

func (set *ProviderSet) Append(provider ...Provider) *ProviderSet {
	set.providers = append(set.providers, provider...)
	return set
}

func (set *ProviderSet) AppendQuery(subtitleQueries ...string) *ProviderSet {
	set.subtitleQueries = append(set.subtitleQueries, subtitleQueries...)
	return set
}

func (set *ProviderSet) StartSearchAndDownload() error {
outer:
	for _, query := range set.subtitleQueries {
		fmt.Println()
		set.options.Query = query
		pterm.Info.Printf("Searching Subtitles for: %s\n", set.options.defaultOutputFileName())
		for _, provider := range set.providers {
			pterm.Info.Printf("Trying provider: %s\n", provider.Name())

			err := set.searchSubtitleWithProvider(provider)
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

func (set *ProviderSet) searchSubtitleWithProvider(provider Provider) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := provider.SearchSubtitle(ctx, set.options)
	if err != nil {
		return err
	}

	if set.options.AutoSelect {
		subFile, dlErr := provider.DownloadBest(ctx)
		if dlErr != nil {
			return dlErr
		}
		return set.saveSubtitle(&subFile)
	}

	selected, err := provider.PromptSelection()
	if err != nil {
		return err
	}

	for _, sub := range selected {
		subFile, err := provider.Download(ctx, sub)
		if err != nil {
			return err
		}

		err = set.saveSubtitle(&subFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (set *ProviderSet) saveSubtitle(subFile *SubtitleFile) error {
	defer subFile.Close()

	subFormatter, err := subformat.NewSubFormatter(subFile.Type, subFile)
	if err != nil {
		return err
	}
	opts := set.options
	var outFileName string
	if opts.OutputFile != "" {
		outFileName = util.StripExtension(opts.OutputFile)
	} else {
		outFileName = cmp.Or(subFile.Name, opts.defaultOutputFileName())
	}

	outFileName = subformat.AddExtensionToSubFile(outFileName, opts.SubtitleFormat)
	outFileName = path.Join(opts.OutputDir, outFileName)

	if opts.OutputDir != "" {
		err = util.CreateDirIfNotExists(opts.OutputDir)
		if err != nil {
			return err
		}
	}

	outFile, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = subFormatter.ConvertToAndWrite(opts.SubtitleFormat, outFile)
	if err != nil {
		return err
	}
	pterm.Info.Println("Subtitle saved at: ", outFileName)
	return nil
}

func (o *Options) defaultOutputFileName() string {
	if o.IsSerie {
		sb := strings.Builder{}
		sb.WriteString(o.Query)
		if o.Season != 0 {
			fmt.Fprintf(&sb, " Season %v", o.Season)
		}
		if o.Episode != 0 {
			fmt.Fprintf(&sb, " Episode %v", o.Episode)
		}
		return sb.String()
	}

	return o.Query
}
