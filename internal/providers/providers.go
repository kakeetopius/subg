// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"cmp"
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

type Provider interface {
	Name() string

	WithOptions(Options)

	SearchSubtitle() error

	DisplaySelections() ([]Subtitle, error)

	Download(Subtitle) (string, io.ReadCloser, subformat.FormatType, error)

	DownloadBest() error
}

type Subtitle any

type ProviderSet struct {
	providers       []Provider
	options         Options
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

func NewProviderSet(subtitleQueries []string) ProviderSet {
	return ProviderSet{
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
	provider.WithOptions(set.options)

	err := provider.SearchSubtitle()
	if err != nil {
		return err
	}

	if set.options.AutoSelect {
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

	for _, sub := range selected {
		err := set.downloadAndSaveSubtitle(sub, provider)
		if err != nil {
			return err
		}
	}

	return nil
}

func (set *ProviderSet) downloadAndSaveSubtitle(subtitle Subtitle, provider Provider) error {
	name, subBytes, downloadedFormat, err := provider.Download(subtitle)
	if err != nil {
		return err
	}
	defer subBytes.Close()

	subFormatter, err := subformat.NewSubFormatter(downloadedFormat, subBytes)
	if err != nil {
		return err
	}

	opts := set.options
	var outFileName string
	if opts.OutputFile != "" {
		outFileName = util.StripExtension(opts.OutputFile)
	} else {
		outFileName = cmp.Or(name, opts.defaultOutputFileName())
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
