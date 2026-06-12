package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kakeetopius/subg/internal/formats"
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
			formatter, err := formats.NewSubFormatFromFile(inFile)
			if err != nil {
				return err
			}

			var formatType formats.FormatType
			if outFile != "" {
				outFileExt := filepath.Ext(outFile)
				fType, ferr := formats.SubFormatTypeFromString(outFileExt)
				if ferr != nil {
					return ferr
				}
				formatType = fType
			} else if convertTo != "" {
				fType, ferr := formats.SubFormatTypeFromString(convertTo)
				if ferr != nil {
					return ferr
				}
				formatType = fType
			} else {
				return fmt.Errorf("could not determine which format to convert to.\nprovide either an output file name or use the --convert-to flag")
			}

			if outFile == "" {
				outFile = util.AddSubFileExtension(inFile, formatType)
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
	convertCmd.Flags().StringVarP(&convertTo, "convert-to", "c", "", "The format to convert to (ignored if output file name is given)")

	return &convertCmd
}
