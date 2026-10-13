package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kakeetopius/subg/internal/subformat"
	"github.com/kakeetopius/subg/internal/util"
	"github.com/spf13/cobra"
)

func ConvertCmd() *cobra.Command {
	var (
		outFile   string
		convertTo string
	)

	convertCmd := cobra.Command{
		Use:   "convert file",
		Short: "Convert subtitle from one format to another",
		Long: `Convert subtitle from one format to another.

Supported formats for now are: srt, vtt, ass, ssa, ttml, stl`,
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			file := args[0]
			formatter, err := subformat.NewSubFormatterFromFile(file)
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
				outFile = util.StripExtension(file)
				outFile = subformat.AddExtensionToSubFile(outFile, formatType)
			}

			innerErr := util.CreateDirIfNotExists(filepath.Dir(outFile))
			if innerErr != nil {
				return innerErr
			}

			outFile, innerErr := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
			if innerErr != nil {
				return innerErr
			}
			defer outFile.Close()

			innerErr = formatter.ConvertToAndWrite(formatType, outFile)
			if innerErr != nil {
				return innerErr
			}
			return nil
		},
	}

	convertCmd.Flags().SortFlags = false

	convertCmd.Flags().StringVarP(&outFile, "out", "o", "", "The output file name.")
	convertCmd.Flags().StringVarP(&convertTo, "convert-to", "c", "", "The format to convert to")

	return &convertCmd
}
