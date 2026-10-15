package downloader

import (
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/subformat"
)

func New(set providers.ProviderSet, opts Options) Downloader {
	return Downloader{
		Options:     opts,
		ProviderSet: set,
	}
}

type Downloader struct {
	ProviderSet providers.ProviderSet
	Options
	Queries []providers.SearchOptions
}

type Options struct {
	AutoSelect            bool
	Directory             string
	DefaultEmbedSubtitles bool
	DefaultSubFormat      subformat.FormatType
}

func (d *Downloader) AppendQuery(q ...providers.SearchOptions) {
	d.Queries = append(d.Queries, q...)
}
