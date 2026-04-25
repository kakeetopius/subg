// Package providers contains functions to interact with different subtitle providers.
package providers

type SubtitleSearchResult interface {
	SubtitleByID(id string) (any, error)
}
