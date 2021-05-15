.PHONY: clean build watch test

help:
	@echo "usage: make <command>"
	@echo ""
	@echo "commands:"
	@echo "  clean      remove unwanted stuff"
	@echo "  build      build the app"
	@echo "  watch      run watcher"
	@echo "  test       run the testsuite"
	@echo "  help       display the help message"


clean:
	find . -name 'tmp' -exec rm -rf {} +
	find . -name '*~' -exec rm -f {} +
	find . -name 'main' -exec rm -f {} +
	find . -name 'repack' -exec rm -f {} +
	go clean


build:
	go build


watch:
	air -c .air.toml


test:
	go test
