// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/dreulavelle/jhin/parser"
	"github.com/kakeetopius/subg/internal/embed"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/ui"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

// NewProviderSet creates a new ProviderSet
func NewProviderSet(opts ProviderSetOptions) *ProviderSet {
	return &ProviderSet{
		providers:          make([]Provider, 0, 5),
		ProviderSetOptions: opts,
	}
}

// Append adds one or more providers to the ProviderSet and returns the updated set.
func (set *ProviderSet) Append(provider ...Provider) *ProviderSet {
	set.providers = append(set.providers, provider...)
	return set
}

// AppendQuery adds one or more subtitle search queries to the ProviderSet and returns the updated set.
func (set *ProviderSet) AppendQuery(query ...SearchOptions) *ProviderSet {
	set.Queries = append(set.Queries, query...)
	return set
}

// StartSearchAndDownload starts the search and download process for all subtitle queries in the ProviderSet. It iterates through each query and attempts to find and download subtitles from the registered providers. If a provider returns an error, it will continue to the next provider.
func (set *ProviderSet) StartSearchAndDownload() error {
	if set.Directory != "" {
		queries, err := set.getQueriesFromDirectory(set.Directory)
		if err != nil {
			return err
		}
		set.AppendQuery(queries...)
	}

	if len(set.Queries) == 0 {
		return fmt.Errorf(" No queries provided ")
	}

outer:
	for _, query := range set.Queries {
		if query.InputFileName != "" {
			parsedQuery := parseFileName(query.InputFileName, &query) // if an input file was given, parse it to get structured query info from it
			query = parsedQuery
		}

		fmt.Println()
		pterm.Info.Printf("Searching Subtitles for: %s\n", query.defaultName())

		for i, provider := range set.providers {
			pterm.Info.Printf("Trying provider: %s\n", provider.Name())

			err := set.searchSubtitleWithProvider(provider, query)
			if err == nil {
				continue outer // move on to the next query
			}

			if errors.Is(err, ui.ErrUserQuit) {
				return nil
			}

			if errors.Is(err, ui.ErrNextProvider) && i < len(set.providers) {
				continue // if next provider is requested and there are still some providers
			}

			if !errors.Is(err, ui.ErrNextProvider) { // prevents printing of ErrNextProvider
				pterm.Error.Printf("Provider '%v' returned an error: %v\n", provider.Name(), err)
			}

			if i == len(set.providers)-1 {
				pterm.Info.Printf("No more providers to try\n")
			}
		}
	}

	return nil
}

func (set *ProviderSet) getQueriesFromDirectory(dir string) ([]SearchOptions, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	searches := make([]SearchOptions, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !util.IsVideoExtension(util.ExtensionOf(file.Name())) {
			continue
		}
		searchOpt := parseFileName(file.Name(), nil)
		searchOpt.OutputDir = dir
		searchOpt.Embed = set.DefaultEmbedSubtitles

		searches = append(searches, searchOpt)
	}

	if len(searches) == 0 {
		pterm.Warning.Println("Could not find any video files in the directory: ", dir)
	}

	return searches, nil
}

// searchSubtitleWithProvider searches for subtitles using the specified provider and query parameters.
func (set *ProviderSet) searchSubtitleWithProvider(provider Provider, queryParams SearchOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spinner, err := pterm.DefaultSpinner.Start("Searching subtitles on ", provider.Name())
	if err != nil {
		return err
	}
	defer spinner.Stop()

	subs, err := provider.SearchSubtitle(ctx, queryParams)
	if err != nil {
		spinner.Fail()
		return err
	}

	if len(subs) == 0 {
		spinner.Fail()
		return fmt.Errorf("no subtitles returned by provider '%v", provider.Name())
	}

	spinner.Success("Search Done")

	// if user wants to autoselect we try to download all the subs returned by the provider until one is successful.
	if set.AutoSelect {
		for i, sub := range subs {
			err = set.tryDownloadSub(ctx, sub, provider, &queryParams)
			if err == nil {
				return nil
			}
			pterm.Error.Printf("Error downloading subtitle: %v\n", err)
			if i < len(subs)-1 {
				pterm.Info.Println("Trying the next subtitle returned by the provider.")
			}
		}
		// try next provider coz all the subs returned by this provider failed.
		return ui.ErrNextProvider
	}

	// if user wants to select subtitles for themselves we prompt for selection(s) and download each and returning an error as soon as one fails.
	selected, err := promptSelection(provider.SubtitleHeaders(), subs)
	if err != nil {
		return err
	}

	for _, sub := range selected {
		err = set.tryDownloadSub(ctx, sub, provider, &queryParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (set *ProviderSet) tryDownloadSub(ctx context.Context, sub Subtitle, provider Provider, queryParams *SearchOptions) error {
	spinner, err := pterm.DefaultSpinner.Start("Downloading Subtitle")

	defer func() {
		if err != nil {
			spinner.Fail()
		}
	}()

	subFile, err := provider.Download(ctx, sub)
	if err != nil {
		return err
	}
	spinner.Success("Download Complete")

	if queryParams.Embed && queryParams.InputFileName != "" {
		var subStream io.Reader = subFile.ReadCloser
		if queryParams.KeepSubAfterEmbed {
			// if we need to keep subtitle after embeding we copy the subtitle into a new buffer that will then be used for embedding.
			newSubStream := bytes.NewBuffer(make([]byte, 1024))
			_, err = io.Copy(newSubStream, subFile.ReadCloser)
			if err != nil {
				return err
			}
			subStream = newSubStream
		}

		err = embed.EmbedSubtitle(ctx, embed.EmbedOptions{
			File:      filepath.Join(queryParams.OutputDir, queryParams.InputFileName),
			Subtitle:  subStream,
			Overwrite: true,
		})
		if err != nil {
			return fmt.Errorf("error embedding subtitle: %w", err)
		}
	}

	if queryParams.Embed && !queryParams.KeepSubAfterEmbed {
		return nil
	}

	outFileName, err := set.saveSubtitle(&subFile, queryParams)
	if err != nil {
		return err
	}

	pterm.Info.Println("Subtitle saved at: ", outFileName)
	return nil
}

// saveSubtitle saves the downloaded subtitle file to the specified output directory and format and returns the path where it was saved or an error.
func (set *ProviderSet) saveSubtitle(subFile *SubtitleFile, queryParams *SearchOptions) (string, error) {
	defer subFile.Close()

	subFormatter, err := subformat.NewSubFormatter(subFile.Type, subFile)
	if err != nil {
		return "", err
	}
	var outFileName string
	if queryParams.OutputFile != "" {
		outFileName = util.StripExtension(queryParams.OutputFile)
	} else {
		outFileName = cmp.Or(subFile.Name, queryParams.defaultName())
	}

	outFileName = subformat.AddExtensionToSubFile(outFileName, queryParams.SubtitleFormat)
	outFileName = path.Join(queryParams.OutputDir, filepath.Base(outFileName))

	if queryParams.OutputDir != "" {
		err = util.CreateDirIfNotExists(queryParams.OutputDir)
		if err != nil {
			return "", err
		}
	}

	outFile, err := os.OpenFile(outFileName, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	err = subFormatter.ConvertToAndWrite(queryParams.SubtitleFormat, outFile)
	if err != nil {
		return "", err
	}
	return outFileName, nil
}

// defaultName returns a default output file name based on the search options. If the search is for a series, it includes the season and episode numbers in the file name.
func (o *SearchOptions) defaultName() string {
	sb := strings.Builder{}
	sb.WriteString(o.Query)
	if o.Season != 0 {
		fmt.Fprintf(&sb, " Season %v", o.Season)
	}
	if o.Episode != 0 {
		fmt.Fprintf(&sb, " Episode %v", o.Episode)
	}
	return sb.String()
}

// promptSelection displays a table of subtitles and prompts the user to select one. It returns the selected subtitle(s) or an error if the selection process fails.
func promptSelection(columnHeaders []string, subs []Subtitle) ([]Subtitle, error) {
	columnMaxWidths := make([]int, len(columnHeaders)) // the maximum width for each column

	rows := []table.Row{}
	for _, sub := range subs {
		fields := sub.Fields()
		if len(fields) != len(columnHeaders) {
			return nil, fmt.Errorf("invalid number of fields returned by subtitle of ID: %s. Wanted %v (len %d). Got %v (len %d)", sub.ID(), columnHeaders, len(columnHeaders), fields, len(fields))
		}

		row := make([]string, 0, len(fields))
		for i, field := range fields {
			if len(field) > columnMaxWidths[i] {
				columnMaxWidths[i] = len(field)
			}
			row = append(row, field)
		}

		rows = append(rows, row)
	}

	columns := make([]table.Column, 0, len(columnHeaders))
	for i, header := range columnHeaders {
		if len(header) > columnMaxWidths[i] {
			columnMaxWidths[i] = len(header)
		}

		columns = append(columns, table.Column{
			Title: header,
			Width: columnMaxWidths[i],
		})
	}

	subID, err := ui.DisplayTableAndGetSelectedID(rows, columns)
	if err != nil {
		return nil, err
	}

	sub, err := subtitleByID(subs, subID)
	if err != nil {
		return nil, err
	}

	return []Subtitle{sub}, nil
}

// subtitleByID returns the subtitle with the specified ID from the given slice of subtitles.
func subtitleByID(subs []Subtitle, id string) (Subtitle, error) {
	for _, sub := range subs {
		if sub.ID() == id {
			return sub, nil
		}
	}

	return nil, fmt.Errorf("subtitle with id %v not found in results", id)
}

// parseFileName parses filenames for movies and tv shows into meaningful structured data that can be used for searching and combines 'old' to give a new SearchOptions struct.
func parseFileName(fileName string, old *SearchOptions) SearchOptions {
	var opts SearchOptions

	result := parser.Parse(filepath.Base(fileName))
	opts.Query = result.Title
	opts.InputFileName = fileName

	if len(result.Episodes) != 0 {
		opts.Episode = result.Episodes[0]
		opts.IsSerie = true
	}

	if len(result.Seasons) != 0 {
		opts.Season = result.Seasons[0]
		opts.IsSerie = true
	}

	if result.Year != "" {
		year, err := strconv.Atoi(result.Year)
		if err == nil {
			opts.Year = year
		}
	}

	if old != nil {
		opts.Embed = old.Embed
		opts.IMDBId = old.IMDBId
		opts.Language = old.Language
		opts.Type = old.Type
		opts.SubtitleFormat = old.SubtitleFormat
		opts.OutputFile = old.OutputFile
		opts.OutputDir = old.OutputDir

		if opts.Episode == 0 {
			opts.Episode = old.Episode
		}
		if opts.Season == 0 {
			opts.Season = old.Season
		}
		if opts.Year == 0 {
			opts.Year = old.Year
		}
	}

	return opts
}
