package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/sh4nks/repack/app"
	"github.com/sh4nks/repack/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	dryRun    bool
	noColor   bool
	formats   []string
	force     bool
	verbosity int
)
var ErrNoInputPath error = errors.New("Input path does not exist")
var ErrOutputPathAlreadyExists error = errors.New("Output path already exists.")

var rootCmd = &cobra.Command{
	Use:     "repack [flags] INPUT_DIR [OUTPUT_DIR]",
	Short:   "A CLI tool for repacking cbr and cbz archives",
	Version: makeVersionString(),
	Args:    cobra.RangeArgs(1, 2),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setting up logger
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: noColor}
		output.FormatLevel = func(i interface{}) string {
			return utils.ColorizedFormatLevel(i, noColor)
		}
		log.Logger = log.Output(output)

		switch verbosity {
		case 0:
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case 1:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case 2:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		default:
			fmt.Println("Invalid verbosity level. Valid values are 0-2.")
			os.Exit(1)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRun {
			log.Info().Msg("Running in dry-run mode - no archives will be repacked")
		}

		input_dir, _ := filepath.Abs(args[0])
		if !utils.PathExists(input_dir) {
			return ErrNoInputPath
		}

		var output_dir string
		if len(args) == 2 {
			output_dir, _ = filepath.Abs(args[1])
		} else {
			if args[0] == "." {
				output_dir = filepath.Join(input_dir, "repacked")
			} else {
				output_dir = filepath.Join(filepath.Dir(input_dir), "repacked", filepath.Base(input_dir))
			}
		}

		log.Info().Msgf("Using  INPUT_DIR: %s", input_dir)
		log.Info().Msgf("Using OUTPUT_DIR: %s", output_dir)
		if utils.PathExists(output_dir) {
			if !force {
				return ErrOutputPathAlreadyExists
			}
			log.Warn().Msgf("Force: Overwriting all files in %s", output_dir)
		}

		app, err := app.New(input_dir, output_dir, formats, force)
		if err != nil {
			return err
		}

		app.Run(dryRun)
		return nil
	},
}

func makeVersionString() string {
	return fmt.Sprintf("%s built with %s", app.Version, runtime.Version())
}

func updateVersionTemplate(cmd *cobra.Command) {
	rootCmd.SetVersionTemplate(
		`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`,
	)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "If set, no archives will be repacked")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "If set, no colors are used for the log message")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Forcefully create the archive by deleting and overwriting any existing files and archives")
	rootCmd.PersistentFlags().StringSliceVarP(&formats, "formats", "f", []string{"cbr"}, "The input formats. Multiple formats are supported - comma separated.")
	rootCmd.PersistentFlags().IntVar(&verbosity, "verbosity", 1, "verbosity level (0-3). 0 = error, 1 = info, 2 = debug")
	updateVersionTemplate(rootCmd)
}
