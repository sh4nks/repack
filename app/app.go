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

var SupportedFormats []string = []string{"cbr", "cbz"}
var Version string = "v1.0"

type Formats struct {
	Items []string
}

type App struct {
	inputPath  string
	outputPath string
	force      bool
	formats    *Formats
	rar        *archiver.Rar
	zip        *archiver.Zip
}

func New(inputPath string, outputPath string, format []string, force bool) (*App, error) {
	formats, err := checkSupportedFormats(format)

	if err != nil {
		log.Error().Msg("Error while checking the supported archive formats")
		return nil, err
	}

	return &App{
		inputPath:  inputPath,
		outputPath: outputPath,
		force:      force,
		formats:    formats,
		zip:        archiver.NewZip(),
		rar:        archiver.NewRar(),
	}, nil
}

func (app *App) Run(dryRun bool) {
	archives := app.getArchives(app.inputPath)
	if len(archives) == 0 {
		log.Info().Msg("No archives found.")
		return
	}

	if dryRun {
		log.Info().Msg("No further actions executed due to dry-run.")
		return
	}

	for _, path := range archives {
		log.Info().Msgf("Archive: %s", path)

		dirname, err := app.extract(path)
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

func (app *App) extract(archivePath string) (string, error) {
	filename := filepath.Base(strings.TrimSuffix(archivePath, filepath.Ext(archivePath)))
	relToInput, _ := filepath.Rel(app.inputPath, filepath.Dir(archivePath))
	dirname := filepath.Join(app.outputPath, relToInput, filename)

	if utils.PathExists(dirname) {
		if !app.force {
			log.Error().Msgf("Can't extract into existing path: %s", dirname)
			return "", fmt.Errorf("extract: Can't extract into existing path. Use the force Luke!")
		}
		log.Info().Msgf("Deleting existing destination directory %s", dirname)
		os.RemoveAll(dirname)
	}

	log.Info().Msgf("Extracting: %s into %s", archivePath, dirname)
	var err error
	switch app.formats.GetSuffix(archivePath) {
	case "cbr":
		// we extract into the top level folder instead of a folder named like
		// the archive because the archiver is creating a folder named like
		// the archive
		err = app.rar.Unarchive(archivePath, filepath.Dir(dirname))
	case "cbz":
		err = app.zip.Unarchive(archivePath, filepath.Dir(dirname))
	}

	if err != nil {
		return "", err
	}

	return dirname, nil
}

func (app *App) compress(srcPath string) error {
	zipArchive := srcPath + ".zip"
	cbzArchive := srcPath + ".cbz"

	log.Info().Msgf("Compressing: %s", srcPath)
	err := app.zip.Archive([]string{srcPath}, zipArchive)

	if err != nil {
		log.Error().Msgf("Can't create archive: %v", err)
		return fmt.Errorf("compress: %w", err)
	}

	if utils.PathExists(cbzArchive) && !app.force {
		log.Error().Msgf("Can't rename .zip to .cbz because there is already a .cbz file at the destination location: %v", cbzArchive)
		return fmt.Errorf("compress: Can't rename .zip to .cbz")
	} else {
		log.Warn().Msgf("Force: Overwritting .cbz file at the destination location: %v", cbzArchive)
	}

	err = os.Rename(zipArchive, cbzArchive)
	if err != nil {
		log.Error().Msgf("compress: couldn't rename file: %w", err)
		return err
	}

	log.Info().Msgf("Deleting extracted source folder: %s", srcPath)
	err = os.RemoveAll(srcPath)
	if err != nil {
		log.Error().Msgf("compress: couldn't delete source folder: %w", err)
		return err
	}

	return nil
}

func (app *App) clean(srcPath string) error {
	err := filepath.Walk(srcPath,
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

func (app *App) getArchives(path string) []string {
	archives := []string{}
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if app.formats.HasSuffix(path) {
					log.Debug().Msgf("Found archive: %s", path)
					archives = append(archives, path)
				} else {
					log.Debug().Msgf("Skipping file: %s", path)
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

func (f *Formats) HasSuffix(path string) bool {
	for _, item := range f.Items {
		if strings.HasSuffix(path, item) {
			return true
		}
	}
	return false
}

func (f *Formats) GetSuffix(path string) string {
	for _, item := range f.Items {
		if strings.HasSuffix(path, item) {
			return item
		}
	}
	return ""
}

func checkSupportedFormats(formats []string) (*Formats, error) {
	for _, passed := range formats {
		isSupported := false
		for _, supported := range SupportedFormats {

			if passed == supported {
				isSupported = true
				break
			}
		}

		if !isSupported {
			return nil, fmt.Errorf("Format '%s' is not supported!", passed)
		}
	}

	f := &Formats{
		Items: formats,
	}
	return f, nil
}
