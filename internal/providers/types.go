// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"context"
	"errors"
	"io"

	"github.com/kakeetopius/subg/internal/subformat"
)

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

var ErrNextProvider = errors.New("user requested next provider")

type ErrNoResultsFound struct {
	Query   string
	Season  int
	Episode int
}

func (e ErrNoResultsFound) Error() string {
	o := SearchOptions{Query: e.Query, Season: e.Season, Episode: e.Episode}
	return "no results found for " + o.defaultName()
}
