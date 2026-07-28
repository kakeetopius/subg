// Package opensubtitles is used to talk to opensubtitles API via a wrapper.
package opensubtitles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/angelospk/opensubtitles-go"
	"github.com/kakeetopius/subg/internal/providers"
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
		opts.UserName, err = util.GetTermInput("Enter opensubtitles.com username", false)
		if err != nil {
			return err
		}
	}
	if opts.Password == "" {
		opts.Password, err = util.GetTermInput("Enter opensubtitles.com password", true)
		if err != nil {
			return err
		}
	}
	if opts.APIKey == "" {
		opts.APIKey, err = util.GetTermInput("Enter opensubtitles.com API Key", false)
		if err != nil {
			return err
		}
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
		if err == nil {
			spinner.Success("Login Successfull")
		} else {
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
	return nil
}

func (p *OpenSubtitles) searchSubtitle(opts providers.SearchOptions) ([]providers.Subtitle, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("API Key is required to download from opensubtitles")
	}

	// To search from opensubtitles the user only needs an api key no need for login
	client, err := opensubtitles.NewClient(opensubtitles.Config{
		ApiKey:    p.APIKey,
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
	defer func() {
		if err == nil {
			spinner.Success("Search Complete")
		} else {
			spinner.Fail()
		}
	}()

	searchResp, err := client.SearchSubtitles(context.Background(), searchParams)
	if err != nil {
		return nil, err
	}

	subtitles := make([]providers.Subtitle, 0, len(searchResp.Data))
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

func (p *OpenSubtitles) downloadSubtitle(ctx context.Context, subtitle *OSSubtitle) (subtitleFile providers.SubtitleFile, err error) {
	// To download from opensubtitles the user must have already logged in a session info cached.
	client, err := newClientFromCachedConfigs(p.APIKey, p.CacheDir)
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
		if err == nil {
			spinner.Success("Download Complete")
		} else {
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

	return providers.SubtitleFile{
		Name:       file2Download.FileName,
		Type:       subformat.FormatTypeSRT,
		ReadCloser: resp.Body,
	}, nil
}
