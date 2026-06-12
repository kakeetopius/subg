package cmd

import (
	"github.com/kakeetopius/subg/internal/formats"
	"github.com/kakeetopius/subg/internal/generate"
	"github.com/spf13/cobra"
)

func GenCmd() *cobra.Command {
	var (
		verbose        bool
		language       string
		model          string
		subtitleFormat string
		translate      bool
		outputDir      string
	)

	genCmd := cobra.Command{
		Use:     "generate files...",
		Short:   "Generate subtitles from video or audio file",
		Aliases: []string{"g", "gen"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			subFormat, err := formats.SubFormatTypeFromString(subtitleFormat)
			if err != nil {
				return err
			}
			return generate.Transcribe(generate.TransciberOptions{
				CacheDir:       appConfig.GetString("cache_dir"),
				HFToken:        appConfig.GetString("transcriber.hf_token"),
				Verbose:        verbose,
				InputFiles:     args,
				SubtitleFormat: subFormat,
				Translate:      translate,
			})
		},
	}

	genCmd.Flags().SortFlags = false
	genCmd.Flags().BoolVar(&verbose, "verbose", false, "Print extra information of what is going on.")
	genCmd.Flags().StringVarP(&language, "lang", "l", "en", "The language of the input file(s)")
	genCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The format to save the subtitle file(s) as.")
	genCmd.Flags().StringVar(&outputDir, "output-dir", "", "The directory to save the subtitles files to.")
	genCmd.Flags().BoolVarP(&verbose, "translate", "t", false, "Translate the file(s) given to English instead of transcribing")

	genCmd.Flags().String("hf-token", "", "The Hugging Face Access Token to access transcribing models.")
	appConfig.BindPFlag("transcriber.hf_token", genCmd.Flags().Lookup("hf-token"))

	genCmd.Flags().StringVarP(&model, "model", "m", "turbo", "The whisper model to use (tiny, medium, large-v3, turbo etc). See https://github.com/openai/whisper/blob/main/model-card.md for more information.")
	return &genCmd
}
