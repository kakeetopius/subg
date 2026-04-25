// Package subdl is used to search for subtitles from subdl.com
package subdl

import (
	"fmt"

	"github.com/kakeetopius/subg/internal/providers"
)

type SearchParams struct {
	Query           *string `url:"film_name,omitempty"`
	FileName        *string `url:"file_name,omitempty"`
	SubDLID         *int    `url:"sd_id,omitempty"`
	IMDBId          *int    `url:"imdb_id,omitempty"`
	TMDBId          *int    `url:"tmdb_id,omitempty"`
	SeasonNumber    *int    `url:"season_number,omitempty"`
	EpisodeNumber   *int    `url:"episode_number,omitempty"`
	Type            *string `url:"type,omitempty"`
	Year            *int    `url:"year,omitempty"`
	Languages       *string `url:"languages,omitempty"`
	SubsPerPage     *int    `url:"subs_per_page,omitempty"`
	Comment         *string `url:"comment,omitempty"`
	Releases        *int    `url:"releases,omitempty"`
	HearingImpaired *int    `url:"hi,omitempty"`
	FullSeason      *int    `url:"full_season,omitempty"`
	APIKey          string  `url:"api_key,omitempty"`
}

type SubtitleFeature struct {
	Name         string  `json:"name"`
	IMDBId       string  `json:"imdb_id"`
	TMDBId       int     `json:"tmdb_id"`
	Type         string  `json:"type"`
	SubDLId      int     `json:"sd_id"`
	FirstAirDate *string `json:"first_air_date,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Year         int     `json:"year"`
}

type SDSubtitle struct {
	Name            string `json:"name"`
	ID              int    `json:"-"` // not part of subdl api
	ReleaseName     string `json:"release_name"`
	Lang            string `json:"lang"`
	Author          string `json:"author"`
	URL             string `json:"url"`
	SubtitlePage    string `json:"subtitlePage"`
	Season          *int   `json:"season"`
	Episode         *int   `json:"episode"`
	LangCode        string `json:"language"`
	HearingImpaired bool   `json:"hi"`
	EpisodeFrom     *int   `json:"episode_from"`
	EpisodeEnd      *int   `json:"episode_end"`
	FullSeason      bool   `json:"full_season"`
}

type DownloadOptions struct {
	Subtitle   *SDSubtitle
	OutPutFile string
	OutPutDir  string
}

type SearchResults struct {
	Status      bool              `json:"status"`
	Results     []SubtitleFeature `json:"results"`
	Subtitles   []SDSubtitle      `json:"subtitles"`
	TotalPages  int               `json:"totalPages"`
	CurrentPage int               `json:"currentPage"`
}

func (r *SearchResults) SubtitleByID(id string) (providers.Subtitle, error) {
	for _, sub := range r.Subtitles {
		idStr := fmt.Sprint(sub.ID)
		if idStr == id {
			return &sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}

func (s *SDSubtitle) Download(dlOpts any) error {
	var opts DownloadOptions
	var ok bool
	if opts, ok = dlOpts.(DownloadOptions); !ok {
		return fmt.Errorf("wrong download options given")
	}
	opts.Subtitle = s

	return DownloadSubtitle(opts)
}
