// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kakeetopius/subg/internal/subformat"
)

// Provider searches for and downloads subtitles from a subtitle service.
type Provider interface {
	// Name returns the provider's display name.
	Name() string

	// SearchSubtitle searches for subtitles matching the supplied options and returns the matching subtitles.
	SearchSubtitle(ctx context.Context, opts SearchOptions) ([]Subtitle, error)

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

// SearchOptions are the general options for searching and downloading subtitles.
type SearchOptions struct {
	// Search Options
	Query             string
	IMDBId            int
	Season            int
	Episode           int
	Language          string
	Type              string
	Year              int
	IsMovie           bool
	IsSerie           bool
	InputFileName     string
	Embed             bool
	KeepSubAfterEmbed bool

	// Download Options
	SubtitleFormat subformat.FormatType
	OutputFile     string
	OutputDir      string
}

// DefaultName returns a default output file name based on the search options. If the search is for a series, it includes the season and episode numbers in the file name.
func (o *SearchOptions) DefaultName() string {
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

// SubtitleFile represents the contents of a downloaded subtitle.
type SubtitleFile struct {
	// Name is the suggested filename of the subtitle. Can be empty
	Name string

	// Type identifies the subtitle format.
	Type subformat.FormatType

	// ReadCloser provides access to the subtitle contents.
	io.ReadCloser
}

type ProviderSet struct {
	// providers is the set of registered subtitle providers.
	providers       []Provider
	nextProviderIdx int
}

type ErrNoResultsFound struct {
	Query   string
	Season  int
	Episode int
}

func (e ErrNoResultsFound) Error() string {
	o := SearchOptions{Query: e.Query, Season: e.Season, Episode: e.Episode}
	return "no results found for " + o.DefaultName()
}

var ErrNoMoreProviders = errors.New("no more providers in provider set.")
