// Package generate is used to transcribe and translate subtitles from audio or video files.
package generate

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/kakeetopius/subg/internal/embed"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/pterm/pterm"
)

type TransciberOptions struct {
	SubtitleFormat    subformat.FormatType
	InputFiles        []string
	Directory         string
	OutputDir         string
	Language          string
	Translate         bool
	Embed             bool
	KeepSubAfterEmbed bool
	CacheDir          string
	Verbose           bool
	HFToken           string
	Model             string
}

const (
	transcriberName   = "whisper-ctranslate2"
	transcriberEnvDir = "transcriber"
	defaultModelName  = "large-v3-turbo"
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

	if opts.Directory != "" {
		dirFiles, err := getInputFilesFromDirectory(opts.Directory)
		if err != nil {
			return err
		}
		opts.InputFiles = append(opts.InputFiles, dirFiles...)
	}

	if opts.HFToken == "" {
		pterm.Warning.Println("No Hugging Face Token provided. Please set a hugging face token to enable higher rate limits and faster downloads")
	}
	if opts.Directory != "" && opts.OutputDir == "" {
		opts.OutputDir = opts.Directory // if an input directory was given but not out directory, then out dir should be the input directory
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "." // default to cwd
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
	if !opts.Embed {
		return err
	}

	return embedSubs(&opts)
}

func transcribeFile(opts *TransciberOptions) error {
	pterm.Info.Println("Transcribing Audio. This might take a while.....")

	flags := []string{
		"--model", cmp.Or(opts.Model, defaultModelName),
		"--output_format", "srt",
	}

	if opts.Language != "" {
		flags = append(flags, "--language", opts.Language)
	}
	if opts.OutputDir != "" {
		flags = append(flags, "--output_dir", opts.OutputDir)
	}

	args := buildArgs(opts.InputFiles, flags)

	return runTranscriber(opts, args)
}

func translateFiles(opts *TransciberOptions) error {
	pterm.Info.Println("Translating Audio......")

	flags := []string{
		"--model", cmp.Or(opts.Model, defaultModelName),
		"--task", "translate",
		"--output_format", "srt",
	}

	if opts.Language != "" {
		flags = append(flags, "--language", opts.Language)
	}
	if opts.OutputDir != "" {
		flags = append(flags, "--output_dir", opts.OutputDir)
	}

	args := buildArgs(opts.InputFiles, flags)

	return runTranscriber(opts, args)
}

func embedSubs(opts *TransciberOptions) error {
	for _, file := range opts.InputFiles {
		if !util.IsVideoExtension(util.ExtensionOf(file)) {
			continue
		}
		subPath := subtitleNameFromFileName(file, opts.SubtitleFormat)
		subPath = filepath.Join(opts.OutputDir, subPath)

		// subtitle files should be already saved in the opts.OutputDir with the same name as the input file and with the correct subtitle extension.
		subtitleStream, err := os.Open(subPath)
		if err != nil {
			return err
		}
		defer subtitleStream.Close()
		pterm.Info.Println("Embedding subtitles for file: ", file)

		err = embed.EmbedSubtitle(context.Background(), embed.EmbedOptions{
			File:      file,
			Subtitle:  subtitleStream,
			Overwrite: true,
		})
		if err != nil {
			return err
		}

		if !opts.KeepSubAfterEmbed {
			err = os.Remove(subPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runTranscriber(opts *TransciberOptions, args []string) error {
	transcriber := filepath.Join(opts.CacheDir, transcriberEnvDir, getPyEnvBinDir(), transcriberName)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		awaitSignal()
		cancel()
	}()

	cmd := exec.CommandContext(
		ctx,
		transcriber,
		args...,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", "HF_TOKEN", opts.HFToken))

	err := runCmd(cmd, opts.Verbose)
	if err != nil {
		return err
	}

	return convertFilesToGivenFormatAndSave(opts)
}

func convertFilesToGivenFormatAndSave(opts *TransciberOptions) error {
	var err error
	if opts.OutputDir != "" {
		err = util.CreateDirIfNotExists(opts.OutputDir)
		if err != nil {
			return err
		}
	}
	outDir := opts.OutputDir

	for _, file := range opts.InputFiles {
		// generated file in srt format
		inFile := util.StripExtension(file)
		inFile = subformat.AddExtensionToSubFile(inFile, subformat.FormatTypeSRT)
		inFile = filepath.Join(outDir, inFile)

		// final file with correct format
		outFile := util.StripExtension(file)
		outFile = subformat.AddExtensionToSubFile(outFile, opts.SubtitleFormat)
		outFile = filepath.Join(outDir, outFile)

		err := convertFile(inFile, outFile, opts.SubtitleFormat)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertFile(infile string, outfile string, subFormat subformat.FormatType) error {
	if subFormat == subformat.FormatTypeSRT {
		// if it is an srt file no need to convert
		return nil
	}
	f, e := os.Open(infile)
	if e != nil {
		return e
	}
	defer f.Close()

	subFormatter, err := subformat.NewSubFormatter(subformat.FormatTypeSRT, f)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		awaitSignal()
		cancel()
	}()

	pterm.Info.Println("Setting up transcriber Environment.")
	cmd := exec.CommandContext(
		ctx,
		"python", "-m", "venv", transcriberEnvPath,
	)

	return runCmd(cmd, verbose)
}

func installTranscriber(cacheDir string, verbose bool) error {
	transcriberPath := filepath.Join(cacheDir, transcriberEnvDir, getPyEnvBinDir(), transcriberName)

	if util.FileExists(transcriberPath) {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		awaitSignal()
		cancel()
	}()

	pterm.Info.Println("Installing the transcriber......")

	pip := filepath.Join(cacheDir, transcriberEnvDir, getPyEnvBinDir(), "pip")

	cmd := exec.CommandContext(
		ctx,
		pip, "install", transcriberName,
	)

	return runCmd(cmd, verbose)
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

func runCmd(cmd *exec.Cmd, verbose bool) error {
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return err
		}
	}

	return nil
}

func getInputFilesFromDirectory(dirName string) ([]string, error) {
	dirEntries, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	inputFiles := make([]string, 0, len(dirEntries))

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if !util.IsVideoOrAudio(util.ExtensionOf(fileName)) {
			continue
		}
		inputFiles = append(inputFiles, filepath.Join(dirName, fileName))
	}

	return inputFiles, nil
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

func awaitSignal() {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGQUIT)

	<-signalChan
}

func getPyEnvBinDir() string {
	// On windows, the python -m venv command places binaries in the 'Scripts' directory
	if runtime.GOOS == "windows" {
		return "Scripts"
	}

	return "bin"
}

func subtitleNameFromFileName(videoFileName string, subFormat subformat.FormatType) string {
	subFile := subformat.AddExtensionToSubFile(
		util.StripExtension(videoFileName),
		subFormat,
	)
	return filepath.Base(subFile)
}
