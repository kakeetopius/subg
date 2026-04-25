// Package providers contains functions to interact with different subtitle providers.
package providers

type SubtitleSearchResult interface {
	SubtitleByID(id string) (Subtitle, error)

	// BestSubtitle() -> Get best subtitle from result Set
}

type Subtitle interface {
	Download(downloadOptions any) error
}
