ifneq (,$(wildcard ./.env)) # guard against hard stop on missing .env file (e.g. on bootstrap)
include .env
export
endif

# Use `make V=1` to print commands.
$(V).SILENT:

.PHONY: build

# invoke make with VERSION=n.n.n
TAG := -$(VERSION)
ifeq ($(VERSION),)
TAG :=
endif

setup:
		chmod +x ./devtools/setup.sh
		./devtools/setup.sh

# building and testing
build:
		mkdir -p ./build
		go build -o ./build/nhmlg$(TAG) ./cmd/nhmlg

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test

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

# Run a local test pg instance. Assumes you have docker and docker daemon is running
pgup:
		docker compose --env-file ./.env -f ./devtools/compose.yml up -d
pgdown:
		docker compose --env-file ./.env -f ./devtools/compose.yml down

# connect to db / browse data
connect:
		docker compose --env-file ./.env -f ./devtools/compose.yml exec -it pg \
		psql -U $(PG_USER) -d $(PG_DB)

# see list of migrations
ls:
		echo "Migrations directory: $(HOME)/nhmlg_datavol/migrations"
		ls -al $(HOME)/nhmlg_datavol/migrations

clean:
		rm ./build/*

# recycle the local dev db volume (start afresh)
recycle: 
		./devtools/recycle-data-volume.sh

uninstall:
		rm $(GOPATH)/bin/nhml-graph
