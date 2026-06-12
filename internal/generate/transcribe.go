// Package generate is used to transcribe and translate subtitles from audio or video files.
package generate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kakeetopius/subg/internal/formats"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

type TransciberOptions struct {
	SubtitleFormat formats.FormatType

	InputFiles []string
	OutPutDir  string
	Language   string
	Translate  bool

	CacheDir string
	Verbose  bool

	HFToken string
	Model   string
}

const (
	transcriberName   = "whisper-ctranslate2"
	transcriberEnvDir = "transcriber"
)

func Transcribe(opts TransciberOptions) error {
	err := checkForBinaryDependencies()
	if err != nil {
		return err
	}

	err = validateInputFiles(opts.InputFiles)
	if err != nil {
		return err
	}

	if opts.HFToken == "" {
		pterm.Warning.Println("No Hugging Face Token provided. Please set a hugging face token to enable higher rate limits and faster downloads")
	}

	err = initTranscirberPyEnv(opts.CacheDir, opts.Verbose)
	if err != nil {
		return err
	}

	err = installTranscriber(opts.CacheDir, opts.Verbose)
	if err != nil {
		return err
	}

	if opts.Verbose {
		pterm.Info.Printf("Using model: %s\n", opts.Model)
	}

	if opts.Translate {
		err = translateFiles(&opts)
	} else {
		err = transcribeFile(&opts)
	}
	if err != nil {
		return err
	}
	return nil
}

func transcribeFile(opts *TransciberOptions) error {
	pterm.Info.Println("Transcribing Audio. This might take a while if the model is not yet offline.....")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := buildArgs(opts.InputFiles, []string{
		"--model", defaultVal(opts.Model, "turbo"),
		"--language", defaultVal(opts.Language, "en"),
		"--output_dir", defaultVal(opts.OutPutDir, cwd),
		"--output_format", "srt",
	})

	return runTranscriber(opts, args)
}

func translateFiles(opts *TransciberOptions) error {
	pterm.Info.Println("Translating Audio......")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := buildArgs(opts.InputFiles, []string{
		"--model", defaultVal(opts.Model, "turbo"),
		"--task", "translate",
		"--language", defaultVal(opts.Language, "en"),
		"--output_dir", defaultVal(opts.OutPutDir, cwd),
		"--output_format", "srt",
	})

	return runTranscriber(opts, args)
}

func runTranscriber(opts *TransciberOptions, args []string) error {
	transcriber := filepath.Join(opts.CacheDir, transcriberEnvDir, "bin", transcriberName)
	cmd := exec.CommandContext(
		context.Background(),
		transcriber,
		args...,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", "HF_TOKEN", opts.HFToken))

	if opts.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	if opts.SubtitleFormat != formats.FormatTypeSRT {
		return convertFilesToGivenFormat(opts)
	}

	return nil
}

func convertFilesToGivenFormat(opts *TransciberOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	outDir := defaultVal(opts.OutPutDir, cwd)

	for _, file := range opts.InputFiles {
		file = util.StripExtension(file)

		inFile := filepath.Join(outDir, file)
		outFile := filepath.Join(outDir, file)

		inFile = util.AddSubFileExtension(inFile, formats.FormatTypeSRT)
		outFile = util.AddSubFileExtension(outFile, opts.SubtitleFormat)

		err := convertFile(inFile, outFile, opts.SubtitleFormat)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertFile(infile string, outfile string, subFormat formats.FormatType) error {
	f, e := os.Open(infile)
	if e != nil {
		return e
	}
	defer f.Close()

	subFormatter, err := formats.NewSubFormat(formats.FormatTypeSRT, f)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(outfile, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	err = subFormatter.ConvertToAndWrite(subFormat, out)
	if err != nil {
		return err
	}

	return os.Remove(infile)
}

func initTranscirberPyEnv(cacheDir string, verbose bool) error {
	transcriberEnvPath := filepath.Join(cacheDir, transcriberEnvDir)

	if util.DirExists(transcriberEnvPath) {
		return nil
	}

	pterm.Info.Println("Setting up transcriber Environment.")
	cmd := exec.CommandContext(
		context.Background(),
		"python", "-m", "venv", transcriberEnvPath,
	)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func installTranscriber(cacheDir string, verbose bool) error {
	transcriberPath := filepath.Join(cacheDir, transcriberEnvDir, "bin", transcriberName)

	if util.FileExists(transcriberPath) {
		return nil
	}

	pterm.Info.Println("Installing the transcriber......")

	pip := filepath.Join(cacheDir, transcriberEnvDir, "bin", "pip")

	cmd := exec.CommandContext(
		context.Background(),
		pip, "install", transcriberName,
	)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func defaultVal(try string, defaultVal string) string {
	if try == "" {
		return defaultVal
	}

	return try
}

func buildArgs(sets ...[]string) []string {
	if len(sets) == 0 {
		return []string{}
	}

	completeArgs := make([]string, 0, len(sets[0]))

	for _, args := range sets {
		completeArgs = append(completeArgs, args...)
	}

	return completeArgs
}

func validateInputFiles(inFiles []string) error {
	for _, file := range inFiles {
		if !util.FileExists(file) {
			return fmt.Errorf(" The file \"%s\" does not exist", file)
		}
	}

	return nil
}

func checkForBinaryDependencies() error {
	_, err := exec.LookPath("python")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("python not found in the $PATH. python is required to run the transcriber. Install it first (or add it to $PATH) and then try to transcribe again. ")
		}
		return err
	}

	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("ffmpeg not found in the $PATH. ffmpeg is required by the transcriber. Install it first (or add it to $PATH) and then try to transcribe again. ")
		}
		return err
	}

	return nil
}
