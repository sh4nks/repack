package app

import (
	"fmt"
	"os"
	"path/filepath"
	"repack/utils"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/rs/zerolog/log"
)

type SupportedFormats []string

var SUPPORTED_FORMATS = SupportedFormats{
	"cbr",
	"cbz",
}

type Archiver interface {
	archiver.Unarchiver
}

type App struct {
	InputPath  string
	OutputPath string
	force      bool
	format     string
	rar        *archiver.Rar
	zip        *archiver.Zip
}

func New(inputPath string, outputPath string, format string, force bool) *App {

	if format == "" || !SUPPORTED_FORMATS.Has(format) {
		log.Error().Msgf("Archive format '%s' not supported.", format)
		return nil
	}

	return &App{
		InputPath:  inputPath,
		OutputPath: outputPath,
		force:      force,
		format:     format,
		zip:        archiver.NewZip(),
		rar:        archiver.NewRar(),
	}
}

func (app *App) Run() {
	archives := getArchives(app.InputPath)
	if len(archives) == 0 {
		log.Info().Msg("No archives found.")
		return
	}
	for _, path := range archives {
		if !strings.HasSuffix(path, app.format) {
			log.Debug().Msgf("Skipping: %v", path)
			continue
		}

		dirname, err := app.extract(path, true)
		if err != nil {
			log.Fatal().Msgf("An error occured during extraction: %v", err)
		}

		app.clean(dirname)

		err = app.compress(dirname)
		if err != nil {
			log.Fatal().Msgf("An error occured during archiving: %v", err)
		}
	}
}

func (app *App) extract(path string, force bool) (string, error) {
	filename := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
	dirname := filepath.Join(filepath.Dir(path), filename)
	var err error

	if utils.PathExists(dirname) {
		if !force {
			log.Error().Msgf("Can't extract into existing path: %s", dirname)
			return "", fmt.Errorf("extract: Can't extract into existing path. Use the force Luke!")
		}
		log.Info().Msgf("Deleting existing destination directory %s", dirname)
		os.RemoveAll(dirname)
	}

	log.Info().Msgf("Extracting: %s", path)
	switch SUPPORTED_FORMATS.GetSuffix(path) {
	case "cbr":
		err = app.rar.Unarchive(path, dirname)
	case "cbz":
		err = app.zip.Unarchive(path, dirname)
	}

	if err != nil {
		return "", err
	}

	return dirname, nil
}

func (app *App) compress(path string) error {
	cbzArchive := filepath.Join(app.OutputPath, path) + ".cbz"
	zipArchive := filepath.Join(app.OutputPath, path) + ".zip"
	log.Info().Msgf("Compressing: %s", path)
	err := app.zip.Archive([]string{path}, zipArchive)

	if err != nil {
		log.Error().Msgf("Can't create archive: %v", err)
		return fmt.Errorf("compress: %w", err)
	}

	if utils.PathExists(cbzArchive) {
		log.Error().Msgf("Can't rename .zip to .cbz because there is already a .cbz file at the destination location: %v", cbzArchive)
		return fmt.Errorf("compress: Can't rename .zip to .cbz")
	}

	err = os.Rename(zipArchive, cbzArchive)
	if err != nil {
		log.Error().Msgf("compress: couldn't rename file: %w", err)
		return err
	}
	return nil
}

func (app *App) clean(path string) error {
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".db") {
				log.Debug().Msgf("Found thumbnail file: %s", path)
				err = os.Remove(path)
				if err != nil {
					log.Error().Msgf("Couldn't delete thumbnail file: %v", err)
					return fmt.Errorf("clean: %w", err)
				}
			}
			return nil
		})

	if err != nil {
		log.Error().Msgf("error: %v\n", err)
		return nil
	}
	return nil
}

func getArchives(path string) []string {
	archives := []string{}
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if SUPPORTED_FORMATS.HasSuffix(path) {
					log.Debug().Msgf("Found archive: %s", path)
					archives = append(archives, path)
				}
			}
			return nil
		})

	if err != nil {
		log.Error().Msgf("error: %v\n", err)
		return nil
	}
	return archives
}

func (formatSlice SupportedFormats) Has(format string) bool {
	for _, item := range formatSlice {
		if item == format {
			return true
		}
	}
	return false
}

func (formatSlice SupportedFormats) HasSuffix(path string) bool {
	for _, item := range formatSlice {
		if strings.HasSuffix(path, item) {
			return true
		}
	}
	return false
}

func (formatSlice SupportedFormats) GetSuffix(path string) string {
	for _, item := range formatSlice {
		if strings.HasSuffix(path, item) {
			return item
		}
	}
	return ""
}
