package cmd

import (
	"github.com/kakeetopius/subg/internal/config"
	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/transcriber"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GenCmd() *cobra.Command {
	var (
		verbose        bool
		language       string
		model          string
		subtitleFormat string
		translate      bool
		directory      string
		outputDir      string
		embed          bool
		keep           bool
	)

	genCmd := cobra.Command{
		Use:     "generate file",
		Short:   "Generate subtitles from video or audio file",
		Aliases: []string{"g", "gen"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			config, err := config.New(configPath(cmd), genCmdConfigSetup(cmd))
			if err != nil {
				return err
			}
			config.Load()

			subFormat, err := subformat.SubFormatFromString(subtitleFormat)
			if err != nil {
				return err
			}
			return transcriber.Transcribe(transcriber.Options{
				CacheDir:          config.CacheDir,
				HFToken:           config.Transcriber.HFToken,
				Verbose:           verbose,
				InputFiles:        args,
				Directory:         directory,
				SubtitleFormat:    subFormat,
				Translate:         translate,
				OutputDir:         outputDir,
				Language:          language,
				Embed:             embed,
				KeepSubAfterEmbed: keep,
				Model:             model,
			})
		},
	}

	genCmd.Flags().SortFlags = false
	genCmd.Flags().BoolVar(&verbose, "verbose", true, "Print extra information of what is going on.")
	genCmd.Flags().StringVarP(&language, "language", "l", "", "The language of the input file(s)")
	genCmd.Flags().StringVarP(&subtitleFormat, "format", "f", "srt", "The format to save the subtitle file(s) as.")
	genCmd.Flags().StringVarP(&directory, "dir", "d", "", "Generate subtitles for all video/audio files in a directory.")
	genCmd.Flags().StringVar(&outputDir, "output-dir", "", "The directory to save the subtitles files to.")
	genCmd.Flags().BoolVarP(&translate, "translate", "t", false, "Translate the file(s) given to English instead of transcribing")
	genCmd.Flags().BoolVarP(&embed, "embed", "e", false, "Embed generated subtitle into video. (If the input is a video file.)")
	genCmd.Flags().BoolVarP(&keep, "keep", "k", false, "Keep subtitle file even after embedding.")

	genCmd.Flags().String("hf-token", "", "The Hugging Face Access Token to access transcribing models.")

	genCmd.Flags().StringVarP(&model, "model", "m", "large-v3-turbo", "The whisper model to use (tiny, medium, large-v3, turbo etc). See https://github.com/openai/whisper/blob/main/model-card.md for more information.")
	return &genCmd
}

func genCmdConfigSetup(cmd *cobra.Command) config.ConfigSetupFunc {
	return func(cfg *viper.Viper) {
		rootCmdConfigSetup(cmd)(cfg)

		err := cfg.BindPFlag("transcriber.hf_token", cmd.Flags().Lookup("hf-token"))
		cobra.CheckErr(err)
		err = cfg.BindEnv("transcriber.hf_token", "HF_TOKEN")
		cobra.CheckErr(err)
	}
}
