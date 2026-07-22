package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/spf13/cobra"
)

func ConvertCmd() *cobra.Command {
	var (
		inFile    string
		outFile   string
		convertTo string
	)

	convertCmd := cobra.Command{
		Use:   "convert",
		Short: "Convert subtitle from one format to another",
		Long: `Convert subtitle from one format to another.

Supported formats for now are: srt, vtt, ass, ssa, ttml, stl`,
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			formatter, err := subformat.NewSubFormatterFromFile(inFile)
			if err != nil {
				return err
			}

			formatType, err := subformat.SubFormatFromFileNameOrFormatString(outFile, convertTo)
			if err != nil {
				if errors.Is(err, subformat.ErrCouldNotDetermineFormat) {
					return fmt.Errorf("could not determine which format to convert to. use either the --convert-to flag or give an output file with the correct extension")
				}
				return err
			}

			if outFile == "" {
				outFile = util.StripExtension(inFile)
				outFile = subformat.AddExtensionToSubFile(outFile, formatType)
			}

			outFile, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0o644)
			if err != nil {
				return err
			}
			defer outFile.Close()

			err = formatter.ConvertToAndWrite(formatType, outFile)
			if err != nil {
				return err
			}
			return nil
		},
	}

	convertCmd.Flags().SortFlags = false
	convertCmd.Flags().StringVarP(&inFile, "in", "i", "", "The input subtitle file.")
	convertCmd.MarkFlagRequired("in")
	convertCmd.MarkFlagFilename("in", "srt", "vtt", "ass", "ssa", "ttml", "stl")

	convertCmd.Flags().StringVarP(&outFile, "out", "o", "", "The output file name.")
	convertCmd.Flags().StringVarP(&convertTo, "convert-to", "c", "", "The format to convert to")

	return &convertCmd
}
