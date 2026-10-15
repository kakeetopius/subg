// Package embed embeds subtitles into video files.
package embed

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

type Options struct {
	File           string
	Subtitle       io.Reader
	OutputFileName string
	Verbose        bool
	Overwrite      bool
}

func EmbedSubtitle(ctx context.Context, opts Options) (err error) {
	err = checkForBinaryDependencies()
	if err != nil {
		return err
	}

	err = validateInputFile(opts.File)
	if err != nil {
		return err
	}

	subCodec, err := subtitleCodec(opts.File)
	if err != nil {
		return err
	}

	if util.Exists(opts.File) && opts.OutputFileName == "" && !opts.Overwrite {
		return fmt.Errorf("The file '%v' already exists and cannot be overwritten.", opts.File)
	}

	outFileSuffix := ".tmp"
	fileExtension := util.ExtensionOf(opts.File)

	// uses the OutputFileName if given otherwise uses the input file name with .tmp added before the extension
	outFileName := cmp.Or(
		opts.OutputFileName,
		fmt.Sprintf("%s%s%s", util.StripExtension(opts.File), outFileSuffix, fileExtension),
	)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGQUIT)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-i", opts.File,
		"-i", "pipe:0", // subtitle file will be passed via stdin (fd 0)
		"-map", "0", // preserve all streams from the video.
		"-map", "1:s", // only select subtitles from second stream
		"-c", "copy", // copy embeded streams as they are without decoding and reencoding them.
		"-c:s", subCodec, // the codec to use for subtitle stream
		outFileName,
	)
	cmd.Stdin = opts.Subtitle // pass subtitle to ffmpeg via stdin

	spinner, err := pterm.DefaultSpinner.Start("Embeding Subtitle")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			spinner.Fail()
		}
	}()

	err = util.RunCmd(cmd, opts.Verbose)
	if err != nil {
		return err
	}
	spinner.Success("Embedding Done.")

	if !opts.Overwrite || opts.OutputFileName != "" {
		return nil
	}
	err = util.CreateDirIfNotExists(filepath.Base(outFileName))
	if err != nil {
		return err
	}

	return os.Rename(outFileName, opts.File)
}

func subtitleCodec(output string) (string, error) {
	switch strings.ToLower(util.ExtensionOf(output)) {
	case ".mp4", ".m4v", ".mov":
		return "mov_text", nil
	case ".mkv":
		return "srt", nil
	default:
		return "", fmt.Errorf("unsupported output container: %s",
			filepath.Ext(output))
	}
}

func checkForBinaryDependencies() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffmpeg not found in the $PATH. ffmpeg is required to extract subtitles from a video file. Install it first (or add it to $PATH) and then try again.")
		}
		return err
	}

	_, err = exec.LookPath("ffprobe")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffprobe not found in the $PATH. ffprobe is required find which are subtitles  a video file. Install it first (or add it to $PATH) and then try again.")
		}
		return err
	}
	return nil
}

func validateInputFile(file string) error {
	if !util.FileExists(file) {
		return fmt.Errorf(" The file '%s' does not exist", file)
	}

	if !util.IsVideoExtension(util.ExtensionOf(file)) {
		return fmt.Errorf("The file '%s' is not a video file", file)
	}
	return nil
}

func awaitSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGQUIT)
	<-signalChan
}
