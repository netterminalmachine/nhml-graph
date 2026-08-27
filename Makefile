# Use `make V=1` to print commands.
$(V).SILENT:

.PHONY: build

# invoke make with VERSION=n.n.n
# WARN: any vars here can be overridden by same-named env vars. DO NOT USE THESE FOR LOCAL DEV.
TAG := -$(VERSION)
ifeq ($(VERSION),)
TAG :=
endif

setup:
		chmod +x ./devtools/setup.sh
		./devtools/setup.sh

# Run a local test pg instance. Assumes you have docker and docker daemon is running
pgup:
		docker compose --env-file ./.env -f ./devtools/compose.yml up -d
pgdown:
		docker compose --env-file ./.env -f ./devtools/compose.yml down


# building and testing
build:
		mkdir -p ./build
		go build -o ./build/nhmlg$(TAG) ./cmd/nhmlg
test:
		make clean
		make build
		clear
		echo "\n=== 📜 TEST RESULTS ===\n"
		gotestsum --format testname
testcov:
		go test -coverpkg=./... -coverprofile=coverage.out ./...
		clear
		echo "\n=== 📜 TEST COVERAGE REPORT ===\n"
		go tool cover -func=coverage.out


# see list of migrations
ls:
		ls -al $(HOME)/nhmlg_datavol/migrations


clean:
		rm ./build/*
uninstall:
		rm $(GOPATH)/bin/nhml-graph
