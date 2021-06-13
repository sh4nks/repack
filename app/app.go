package app

import (
	"compress/flate"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sh4nks/repack/utils"

	"github.com/mholt/archiver/v3"
	"github.com/rs/zerolog/log"
)

var SupportedFormats []string = []string{"cbr", "cbz"}
var Version string = "v1.0"

const fsMode fs.FileMode = 0755

type Formats struct {
	Items []string
}

type RepackItem struct {
	// the full path to the source archive
	src string
	// the full path to the destination/extraction directory
	dst string
	// the extension of the filename
	ext string
	// the filename without the file extension
	name string
}

type App struct {
	inputPath  string
	outputPath string
	force      bool
	formats    *Formats
	rar        *archiver.Rar
	zip        *archiver.Zip
	items      []RepackItem
}

func New(inputPath string, outputPath string, format []string, force bool) (*App, error) {
	formats, err := checkSupportedFormats(format)

	if err != nil {
		log.Error().Msg("Error while checking the supported archive formats")
		return nil, err
	}

	zip := &archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		FileMethod:             archiver.Deflate,
		SelectiveCompression:   true,
		OverwriteExisting:      force,
		ImplicitTopLevelFolder: false,
		MkdirAll:               false,
	}

	rar := &archiver.Rar{
		ImplicitTopLevelFolder: false,
		MkdirAll:               false,
	}

	return &App{
		inputPath:  inputPath,
		outputPath: outputPath,
		force:      force,
		formats:    formats,
		zip:        zip,
		rar:        rar,
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

	for _, item := range archives {
		log.Info().Msgf("Archive: %s", item.src)

		err := app.extract(item)
		if err != nil {
			log.Fatal().Msgf("An error occured during extraction: %v", err)
		}

		app.clean(item.dst)

		err = app.compress(item)
		if err != nil {
			log.Fatal().Msgf("An error occured during archiving: %v", err)
		}
	}
}

func (app *App) extract(item RepackItem) error {
	if utils.PathExists(item.dst) {
		if !app.force {
			log.Error().Msgf("Can't extract into existing path: %s", item.dst)
			return fmt.Errorf("extract: Can't extract into existing path.")
		}
		log.Info().Msgf("Deleting existing destination directory %s", item.dst)
		os.RemoveAll(item.dst)
	}

	err := os.MkdirAll(item.dst, fsMode)
	if err != nil {
		return fmt.Errorf("extract: Couldn't create destination directory: %s", item.dst)
	}

	log.Info().Msgf("Extracting: %s into %s", item.src, item.dst)

	switch app.formats.GetSuffix(item.ext) {
	case "cbr":
		err = app.rar.Unarchive(item.src, item.dst)
	case "cbz":
		err = app.zip.Unarchive(item.src, item.dst)
	}

	if err != nil {
		return err
	}

	return nil
}

func (app *App) compress(item RepackItem) error {
	zipArchive := item.dst + ".zip"
	cbzArchive := item.dst + ".cbz"

	log.Info().Msgf("Compressing: %s", item.dst)

	// Handle top level archive folders
	dirs, _ := os.ReadDir(item.dst)
	var dir string
	if len(dirs) == 1 {
		dir = filepath.Join(item.dst, dirs[0].Name())
	} else {
		dir = item.dst
	}

	err := app.zip.Archive([]string{dir}, zipArchive)

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

	log.Info().Msgf("Deleting extracted source folder: %s", item.dst)
	err = os.RemoveAll(item.dst)
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

func (app *App) getArchives(path string) []RepackItem {
	archives := []RepackItem{}

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if app.formats.HasSuffix(path) {
					log.Debug().Msgf("Found archive: %s", path)

					filename := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
					relToInput, _ := filepath.Rel(app.inputPath, filepath.Dir(path))
					dirname := filepath.Join(app.outputPath, relToInput, filename)

					item := RepackItem{
						src:  path,
						dst:  dirname,
						name: filename,
						ext:  filepath.Ext(path),
					}
					archives = append(archives, item)
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
