package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func RunCmd(cmd *exec.Cmd, verbose bool) error {
	var cmdError bytes.Buffer
	if !verbose && cmd.Stderr == nil {
		// if caller wants verbose=false and hasn't set cmd.Stderr we set it to a buffer so that in case the command fails we print the errors returned.
		cmd.Stderr = &cmdError
	}

	if verbose && cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}

	if verbose && cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		if !verbose {
			// only print the error message if verbose=false, that is when cmdError will contain sth. (if caller didnt set cmd.Stderr)
			fmt.Println(string(cmdError.String()))
		}
		return err
	}

	return nil
}
