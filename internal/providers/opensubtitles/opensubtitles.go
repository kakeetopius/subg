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
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

var CachedCredentialsFile = "auth.json"

type OpenSubLoginOptions struct {
	UserName string
	Password string
	APIKey   string
	CacheDir string
}

type OpenSubSearchOptions struct {
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

type OpenSubDownloadOptions struct {
	Subtitle   *OpenSubSubtitle
	FileID     int
	Format     string
	OutPutFile string
	OutPutDir  string

	APIKey   string
	CacheDir string
}
type OpenSubSubtitle struct {
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

func Login(opts OpenSubLoginOptions) error {
	if opts.UserName == "" {
		return fmt.Errorf("username cannot be empty")
	} else if opts.Password == "" {
		return fmt.Errorf("password cannot be empty")
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
	loginParams := opensubtitles.LoginRequest{
		Username: opts.UserName,
		Password: opts.Password,
	}
	resp, err := client.Login(context.Background(), loginParams)
	if err != nil {
		spinner.Fail()
		return err
	}
	cacheFile, err := util.CreateFileIfNotExists(authFile)
	if err != nil {
		spinner.Fail()
		return err
	}
	defer cacheFile.Close()

	jsonResponse, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		spinner.Fail()
		return err
	}
	_, err = cacheFile.Write(jsonResponse)
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success("Logged in Successfully")
	return nil
}

func SearchSubtitle(opts OpenSubSearchOptions) ([]OpenSubSubtitle, error) {
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
	if opts.SeasonNumber != 0 {
		searchParams.SeasonNumber = &opts.SeasonNumber
	}
	if opts.EpisodeNumber != 0 {
		searchParams.EpisodeNumber = &opts.EpisodeNumber
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

	subtitles := make([]OpenSubSubtitle, 0, len(searchResp.Data))
	for _, sub := range searchResp.Data {
		subtitleObj := OpenSubSubtitle{
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

func NewClientFromCachedConfigs(apiKey string, cacheDir string) (*opensubtitles.Client, error) {
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
	}

	var authResp opensubtitles.LoginResponse
	err = json.Unmarshal(authResponseJSON, &authResp)
	if err != nil {
		return nil, fmt.Errorf("error in auth file %v", err)
	}

	client.SetAuthToken(authResp.Token, authResp.BaseURL)
	return client, nil
}

func DownloadSubtitle(opts OpenSubDownloadOptions) error {
	client, err := NewClientFromCachedConfigs(opts.APIKey, opts.CacheDir)
	if err != nil {
		return err
	}
	if len(opts.Subtitle.Files) == 0 {
		return fmt.Errorf("no files to download for selected subtitle")
	}
	file2Download := opts.Subtitle.Files[0]
	downloadRequest := opensubtitles.DownloadRequest{
		FileID:    file2Download.FileID,
		SubFormat: &opts.Format,
	}
	if opts.OutPutFile == "" {
		opts.OutPutFile = fmt.Sprintf("%v.%v", file2Download.FileName, opts.Format)
	}

	opts.OutPutFile = path.Join(opts.OutPutDir, opts.OutPutFile)
	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle.........")
	if err != nil {
		return err
	}
	downloadResp, err := client.Download(context.Background(), downloadRequest)
	if err != nil {
		spinner.Fail()
		return err
	}

	httpclient := &http.Client{}
	resp, err := httpclient.Get(downloadResp.Link)
	if err != nil {
		spinner.Fail()
		return err
	}
	defer resp.Body.Close()

	outFile, err := os.OpenFile(opts.OutPutFile, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		spinner.Fail()
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		spinner.Fail()
		return err
	}
	spinner.Success("Download Done")

	fmt.Printf("\nSubtitle downloaded successfully to: %v \n", opts.OutPutFile)
	fmt.Printf("Remaining Downloads: %v\n", downloadResp.Remaining)
	fmt.Printf("Reset Time: %v\n", downloadResp.ResetTime)
	return nil
}
