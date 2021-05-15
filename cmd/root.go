package cmd

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/cobra"

	"repack/app"
	"repack/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var rootCmd = &cobra.Command{
	Use:     "repack [flags] INPUT_DIR [OUTPUT_DIR]",
	Version: "v0.1-dev",
	Short:   "A CLI tool for repacking cbr and cbz archives",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No input path provided")
		}

		if len(args) > 2 {
			return errors.New("Unexpected number of arguments")
		}

		if !utils.PathExists(args[0]) {
			return errors.New("Input path does not exist")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msgf("run with args: %v", args)

		var input, output string
		if len(args) == 1 {
			input, output = args[0], args[0]
		} else {
			input, output = args[0], args[1]
		}

		app := app.New(input, output, format, force)
		app.Run()
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

var noColor = false
var format = "cbz"
var force = false
var verbosity = 0

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "If set, no colors are used for the log message")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Forcefully create the archive by deleting and overwriting any existing files and archives")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "cbr", "The input format")
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 1, "verbosity level (0-3). 0 = error, 1 = info, 2 = debug")

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return utils.ColorizedFormatLevel(i, noColor)
	}
	log.Logger = log.Output(output)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
