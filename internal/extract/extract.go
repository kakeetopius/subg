// Package extract is used to extract subtitle files from video file.
package extract

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

type Options struct {
	Files     []string
	Format    subformat.FormatType
	Directory string
	OutputDir string
	Verbose   bool
}

type subtitleStream struct {
	index int
	codec string
}

func ExtractSubtitles(ctx context.Context, opts Options) (err error) {
	err = checkForBinaryDependencies()
	if err != nil {
		return err
	}

	err = validateInputFiles(opts.Files)
	if err != nil {
		return err
	}

	if opts.Directory != "" {
		dirFiles, err := util.GetFilesFromDir(opts.Directory, util.IsVideoExtension)
		if err != nil {
			return err
		}
		opts.Files = append(opts.Files, dirFiles...)
	}

	if opts.Directory != "" && opts.OutputDir == "" {
		opts.OutputDir = opts.Directory // if an input directory was given but not out directory, then out dir should be the input directory
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "." // default to cwd
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGQUIT)
	defer cancel()

	for _, file := range opts.Files {
		subStreams, err := getSubtitleStreamsInFile(ctx, file)
		if err != nil {
			return err
		}
		pterm.Info.Printf("Found %v subtitles in the file: %v\n", len(subStreams), file)
		if len(subStreams) == 0 {
			continue
		}
		err = extractSubs(ctx, subStreams, file, opts.OutputDir, opts.Format)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractSubs(
	ctx context.Context,
	subStreams []subtitleStream,
	fileName string,
	outDir string,
	outformat subformat.FormatType,
) (err error) {
	if outDir != "" {
		err = util.CreateDirIfNotExists(outDir)
		if err != nil {
			return err
		}
	}

	spinner, err := pterm.DefaultSpinner.Start("Extracting")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			spinner.Fail()
		} else {
			spinner.Stop()
		}
	}()

	baseFileName := util.StripExtension(filepath.Base(fileName))

	for i := range subStreams {
		cmd := exec.CommandContext(
			ctx,
			"ffmpeg",
			"-i", fileName,
			"-map", fmt.Sprintf("0:s:%d", i), // the stream to download ie 0:s:i means ith subtitle stream of the first input
			filepath.Join(outDir, subformat.AddExtensionToSubFile(fmt.Sprintf("%s-%d", baseFileName, i+1), outformat)),
		)
		err := util.RunCmd(cmd, false)
		if err != nil {
			return err
		}
	}

	spinner.Success("Extracted subtitles to directory: ", outDir)
	return nil
}

func getSubtitleStreamsInFile(ctx context.Context, fileName string) ([]subtitleStream, error) {
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-select_streams", "s", // only select subtitles
		"-show_entries", "stream=index,codec_name", // only select the subtitle index and codec_name
		"-of", "csv", // return in csv
		fileName,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := util.RunCmd(cmd, false)
	if err != nil {
		return nil, err
	}

	csv := csv.NewReader(&out)
	records, err := csv.ReadAll()
	if err != nil {
		return nil, err
	}

	subStreams := make([]subtitleStream, 0, len(records))

	for _, record := range records {
		if len(record) < 3 {
			continue
		}
		idx, err := strconv.Atoi(record[1])
		if err != nil {
			continue
		}

		subStreams = append(subStreams, subtitleStream{
			index: idx,
			codec: record[2],
		})
	}

	return subStreams, nil
}

func checkForBinaryDependencies() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffmpeg not found in the $PATH. ffmpeg is required to extract subtitles from the video file. Install it first (or add it to $PATH) and then try again.")
		}
		return err
	}

	_, err = exec.LookPath("ffprobe")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffprobe not found in the $PATH. ffprobe is required to find all subtitle streams in the video file. Install it first (or add it to $PATH) and then try again.")
		}
		return err
	}
	return nil
}

func validateInputFiles(files []string) error {
	for _, file := range files {
		if !util.FileExists(file) {
			return fmt.Errorf(" The file '%s' does not exist", file)
		}
		if !util.IsVideoExtension(util.ExtensionOf(file)) {
			return fmt.Errorf("The file '%s' is not a video file", file)
		}
	}

	return nil
}
