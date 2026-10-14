package downloader

import (
	"github.com/kakeetopius/subg/internal/providers"
	"github.com/kakeetopius/subg/internal/subformat"
)

func New(set providers.ProviderSet, opts DownloaderOptions) Downloader {
	return Downloader{
		DownloaderOptions: opts,
		ProviderSet:       set,
	}
}

type Downloader struct {
	ProviderSet providers.ProviderSet
	DownloaderOptions
	Queries []providers.SearchOptions
}

type DownloaderOptions struct {
	AutoSelect            bool
	Directory             string
	DefaultEmbedSubtitles bool
	DefaultSubFormat      subformat.FormatType
}

func (d *Downloader) AppendQuery(q ...providers.SearchOptions) {
	d.Queries = append(d.Queries, q...)
}
