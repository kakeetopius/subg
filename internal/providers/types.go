// Package providers contains functions to interact with different subtitle providers.
package providers

type SubtitleSearchResult interface {
	SubtitleByID(id string) (Subtitle, error)
}

type Subtitle interface {
	Download(downloadOptions any) error
}
