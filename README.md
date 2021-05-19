# repack

A CLI for repacking your cbr and cbz archives in case they seem to be corrupted
or can't be read from your favorite manga reader!


# Usage

```bash
$ repack -h

A CLI tool for repacking cbr and cbz archives

Usage:
  repack [flags] INPUT_DIR [OUTPUT_DIR]

Flags:
      --dry-run           If set, no archives will be repacked
      --force             Forcefully create the archive by deleting and overwriting any existing files and archives
  -f, --formats strings   The input formats. Multiple formats are supported - comma separated. (default [cbr])
  -h, --help              help for repack
      --no-color          If set, no colors are used for the log message
      --verbosity int     verbosity level (0-3). 0 = error, 1 = info, 2 = debug (default 1)
  -v, --version           version for repack
```


# License

repack is licensed under the MIT License.
