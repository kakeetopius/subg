// Package opensubtitles is used to talk to opensubtitles API via a wrapper.
package opensubtitles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/angelospk/opensubtitles-go"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

var CachedCredentialsFile = "auth.json"

type LoginOptions struct {
	UserName string
	Password string
	APIKey   string
	CacheDir string
}

type OSSubtitle struct {
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
	SeasonNumber  int
	EpisodeNumber int
}

type SubtitleFile struct {
	FileID   int
	FileName string
}

func Login(opts LoginOptions) (err error) {
	if opts.UserName == "" {
		return fmt.Errorf("username cannot be empty")
	} else if opts.Password == "" {
		return fmt.Errorf("password cannot be empty")
	} else if opts.APIKey == "" {
		return fmt.Errorf("API Key is required to use opensubtitles")
	}

	authFile := path.Join(opts.CacheDir, CachedCredentialsFile)
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    opts.APIKey,
		UserAgent: "",
	})
	if err != nil {
		return err
	}

	spinner, err := pterm.DefaultSpinner.Start("Logging in.........")
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			spinner.Fail()
		}
	}()

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
	spinner.Success("Logged in Successfully")
	return nil
}

func searchSubtitle(opts options) ([]OSSubtitle, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("API Key is required to download from opensubtitles")
	}

	// To search from opensubtitles the user only needs an api key no need for login
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    opts.APIKey,
		UserAgent: "",
	})
	if err != nil {
		return nil, err
	}
	searchParams := opensubtitles.SearchSubtitlesParams{
		Query:     &opts.Query,
		Languages: &opts.Language,
		Type:      &opts.Type,
	}

	if opts.Year != 0 {
		searchParams.Year = &opts.Year
	}
	if opts.IMDBId != 0 {
		searchParams.IMDbID = &opts.IMDBId
	}
	if opts.Season != 0 {
		searchParams.SeasonNumber = &opts.Season
	}
	if opts.Episode != 0 {
		searchParams.EpisodeNumber = &opts.Episode
	}

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on OpenSubtitles.........")
	if err != nil {
		return nil, err
	}
	searchResp, err := client.SearchSubtitles(context.Background(), searchParams)
	if err != nil {
		spinner.Fail()
		return nil, err
	}
	spinner.Success("Search Done")

	subtitles := make([]OSSubtitle, 0, len(searchResp.Data))
	for _, sub := range searchResp.Data {
		subtitleObj := OSSubtitle{
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

func newClientFromCachedConfigs(apiKey string, cacheDir string) (*opensubtitles.Client, error) {
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    apiKey,
		UserAgent: "",
	})
	if err != nil {
		return nil, err
	}

	authResponseJSON, err := os.ReadFile(path.Join(cacheDir, CachedCredentialsFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("could not find cached opensubtitle credentials. Try subg login --provider os to login to opensubtitles.com")
		}
		return nil, err
	}

	var authResp opensubtitles.LoginResponse
	err = json.Unmarshal(authResponseJSON, &authResp)
	if err != nil {
		return nil, fmt.Errorf("error in auth file %v", err)
	}

	client.SetAuthToken(authResp.Token, authResp.BaseURL)
	return client, nil
}

func downloadSubtitle(opts options, subtitle *OSSubtitle) (name string, sub io.ReadCloser, format subformat.FormatType, err error) {
	// To download from opensubtitles the user must have already logged in a session info cached.
	client, err := newClientFromCachedConfigs(opts.APIKey, opts.CacheDir)
	if err != nil {
		return
	}
	if len(subtitle.Files) == 0 {
		err = fmt.Errorf("no files to download for selected subtitle")
		return
	}
	file2Download := subtitle.Files[0]
	formatString := "srt"

	downloadRequest := opensubtitles.DownloadRequest{
		FileID:    file2Download.FileID,
		SubFormat: &formatString,
	}

	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			spinner.Fail()
		}
	}()

	downloadResp, err := client.Download(context.Background(), downloadRequest)
	if err != nil {
		return
	}

	httpclient := &http.Client{}
	resp, err := httpclient.Get(downloadResp.Link)
	if err != nil {
		return
	}

	return file2Download.FileName, resp.Body, subformat.FormatTypeSRT, nil
}
