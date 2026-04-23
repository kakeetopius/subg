// Package opensubtitles is used to talk to opensubtitles API via a wrapper.
package opensubtitles

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/angelospk/opensubtitles-go"
	"github.com/kakeetopius/subg/internal/util"
)

type LoginOptions struct {
	UserName string
	Password string
	APIKey   string
	CacheDir string
}

type SearchOptions struct {
	Query         string
	IMDBId        int
	SeasonNumber  int
	EpisodeNumber int
	Languages     string
	Type          string
	Year          int

	APIKey   string
	CacheDir string
}

type DownloadOptions struct {
	FileID   int
	Format   string
	FileName string

	APIKey   string
	CacheDir string
}
type Subtitle struct {
	SubtitleID     string
	Release        string
	Votes          int
	Ratings        float64
	UploadDate     time.Time
	URL            string
	Language       string
	FeatureDetails SubtitleFeatureDetails
	Files          []SubtitleFile
}

type SubtitleFeatureDetails struct {
	FeatureID     int
	FeatureType   string
	Year          int
	Title         string
	IMBDId        int
	TMDBId        int
	SeasonNumber  int
	EpisodeNumber int
}

type SubtitleFile struct {
	FileID   int
	FileName string
}

func Login(opts LoginOptions) error {
	if opts.UserName == "" {
		return fmt.Errorf("username cannot be empty")
	} else if opts.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	authFile := path.Join(opts.CacheDir, "auth.json")
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    opts.APIKey,
		UserAgent: "",
	})
	if err != nil {
		return err
	}

	loginParams := opensubtitles.LoginRequest{
		Username: opts.UserName,
		Password: opts.Password,
	}
	resp, err := client.Login(context.Background(), loginParams)
	if err != nil {
		return err
	}
	cacheFile, err := util.CreateFileIfNotExists(authFile)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	jsonResponse, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	_, err = cacheFile.Write(jsonResponse)
	if err != nil {
		return err
	}
	return nil
}

func SearchSubtitle(opts SearchOptions) ([]Subtitle, error) {
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    opts.APIKey,
		UserAgent: "",
	})
	if err != nil {
		return nil, err
	}
	searchParams := opensubtitles.SearchSubtitlesParams{
		Query:     &opts.Query,
		Languages: &opts.Languages,
		Type:      &opts.Type,
	}

	if opts.Year != 0 {
		searchParams.Year = &opts.Year
	}
	if opts.IMDBId != 0 {
		searchParams.IMDbID = &opts.IMDBId
	}
	if opts.Type == "episode" {
		searchParams.SeasonNumber = &opts.SeasonNumber
		searchParams.EpisodeNumber = &opts.EpisodeNumber
	}

	searchResp, err := client.SearchSubtitles(context.Background(), searchParams)
	if err != nil {
		return nil, err
	}

	subtitles := make([]Subtitle, 0, len(searchResp.Data))
	for _, sub := range searchResp.Data {
		subtitleObj := Subtitle{
			SubtitleID: sub.Attributes.SubtitleID,
			Release:    sub.Attributes.Release,
			Votes:      sub.Attributes.Votes,
			Ratings:    sub.Attributes.Ratings,
			UploadDate: sub.Attributes.UploadDate,
			URL:        sub.Attributes.URL,
			Language:   string(sub.Attributes.Language),
			FeatureDetails: SubtitleFeatureDetails{
				FeatureID:   sub.Attributes.FeatureDetails.FeatureID,
				FeatureType: sub.Attributes.FeatureDetails.FeatureType,
				Year:        sub.Attributes.FeatureDetails.Year,
				Title:       sub.Attributes.FeatureDetails.Title,
				IMBDId:      *sub.Attributes.FeatureDetails.IMDbID,
				TMDBId:      *sub.Attributes.FeatureDetails.TMDBID,
			},
		}
		// The following two maybe nil when  dealing with movies
		if sub.Attributes.FeatureDetails.SeasonNumber != nil {
			subtitleObj.FeatureDetails.SeasonNumber = *sub.Attributes.FeatureDetails.SeasonNumber
		}
		if sub.Attributes.FeatureDetails.EpisodeNumber != nil {
			subtitleObj.FeatureDetails.EpisodeNumber = *sub.Attributes.FeatureDetails.EpisodeNumber
		}
		for _, file := range sub.Attributes.Files {
			subtitleObj.Files = append(subtitleObj.Files, SubtitleFile{
				FileID:   file.FileID,
				FileName: file.FileName,
			})
		}

		subtitles = append(subtitles, subtitleObj)
	}
	return subtitles, nil
}
