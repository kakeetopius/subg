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

	"charm.land/bubbles/v2/table"
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

	// SearchSubtitle searches for subtitles matching the supplied options and returns the matching subtitles.
	SearchSubtitle(ctx context.Context, opts SearchOptions) ([]Subtitle, error)

	// SelectBest selects the best subtitle from the given search results according to the provider's selection criteria.
	SelectBest([]Subtitle) Subtitle

	// Download downloads the given subtitle and returns its contents.
	Download(ctx context.Context, subtitle Subtitle) (SubtitleFile, error)

	// SubtitleHeaders returns the column headers used to display subtitle search results in a table. The first column should be a unique subtitle identifier
	SubtitleHeaders() []string
}

// Subtitle represents a provider-specific subtitle search result.
type Subtitle interface {
	// ID returns the unique identifier of the subtitle.
	ID() string

	// Fields returns the values for this subtitle, in the same order as the headers returned by Provider.SubtitleHeaders().
	Fields() []string
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

	// subtitleQueries are the subtitles to search for
	subtitleQueries []SearchOptions
}

// SearchOptions are the general options for searching and downloading subtitles.
type SearchOptions struct {
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

func NewProviderSet() *ProviderSet {
	return &ProviderSet{
		providers: make([]Provider, 0, 5),
	}
}

func (set *ProviderSet) Append(provider ...Provider) *ProviderSet {
	set.providers = append(set.providers, provider...)
	return set
}

func (set *ProviderSet) AppendQuery(query ...SearchOptions) *ProviderSet {
	set.subtitleQueries = append(set.subtitleQueries, query...)
	return set
}

func (set *ProviderSet) StartSearchAndDownload() error {
outer:
	for _, query := range set.subtitleQueries {
		fmt.Println()
		pterm.Info.Printf("Searching Subtitles for: %s\n", query.defaultOutputFileName())
		for _, provider := range set.providers {
			pterm.Info.Printf("Trying provider: %s\n", provider.Name())

			err := set.searchSubtitleWithProvider(provider, query)
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

func (set *ProviderSet) searchSubtitleWithProvider(provider Provider, queryParams SearchOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subs, err := provider.SearchSubtitle(ctx, queryParams)
	if err != nil {
		return err
	}

	todownload := make([]Subtitle, 0, 5)
	if queryParams.AutoSelect {
		todownload = append(todownload, provider.SelectBest(subs))
	} else {
		selected, err := promptSelection(provider.SubtitleHeaders(), subs)
		if err != nil {
			return err
		}
		todownload = append(todownload, selected...)
	}

	for _, sub := range todownload {
		subFile, err := provider.Download(ctx, sub)
		if err != nil {
			return err
		}

		err = set.saveSubtitle(&subFile, &queryParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func (set *ProviderSet) saveSubtitle(subFile *SubtitleFile, queryParams *SearchOptions) error {
	defer subFile.Close()

	subFormatter, err := subformat.NewSubFormatter(subFile.Type, subFile)
	if err != nil {
		return err
	}
	var outFileName string
	if queryParams.OutputFile != "" {
		outFileName = util.StripExtension(queryParams.OutputFile)
	} else {
		outFileName = cmp.Or(subFile.Name, queryParams.defaultOutputFileName())
	}

	outFileName = subformat.AddExtensionToSubFile(outFileName, queryParams.SubtitleFormat)
	outFileName = path.Join(queryParams.OutputDir, outFileName)

	if queryParams.OutputDir != "" {
		err = util.CreateDirIfNotExists(queryParams.OutputDir)
		if err != nil {
			return err
		}
	}

	outFile, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = subFormatter.ConvertToAndWrite(queryParams.SubtitleFormat, outFile)
	if err != nil {
		return err
	}
	pterm.Info.Println("Subtitle saved at: ", outFileName)
	return nil
}

func (o *SearchOptions) defaultOutputFileName() string {
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

func promptSelection(columnHeaders []string, subs []Subtitle) ([]Subtitle, error) {
	columnMaxWidths := make([]int, len(columnHeaders)) // the maximum width for each column

	rows := []table.Row{}
	for _, sub := range subs {
		fields := sub.Fields()
		if len(fields) != len(columnHeaders) {
			return nil, fmt.Errorf("invalid number of fields returned by subtitle of ID: %s. Wanted %v (len %d). Got %v (len %d)", sub.ID(), columnHeaders, len(columnHeaders), fields, len(fields))
		}

		row := make([]string, 0, len(fields))
		for i, field := range fields {
			if len(field) > columnMaxWidths[i] {
				columnMaxWidths[i] = len(field)
			}
			row = append(row, field)
		}

		rows = append(rows, row)
	}

	columns := make([]table.Column, 0, len(columnHeaders))
	for i, header := range columnHeaders {
		if len(header) > columnMaxWidths[i] {
			columnMaxWidths[i] = len(header)
		}

		columns = append(columns, table.Column{
			Title: header,
			Width: columnMaxWidths[i],
		})
	}

	subID, err := ui.DisplayTableAndGetSelectedID(rows, columns)
	if err != nil {
		return nil, err
	}

	sub, err := subtitleByID(subs, subID)
	if err != nil {
		return nil, err
	}

	return []Subtitle{sub}, nil
}

func subtitleByID(subs []Subtitle, id string) (Subtitle, error) {
	for _, sub := range subs {
		if sub.ID() == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}
