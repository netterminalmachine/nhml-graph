#! /bin/bash

# ensure this always runs from the root of the project ($0 is path that launches this script)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "🗲  making sure you have go installed..."
if ! command -v go >/dev/null 2>&1
then
  echo "  ❌ Uh-oh... You don't seem to have go installed."
  echo "Please see here: https://go.dev/doc/install for instructions. After you've installed, you can run this setup again! :-)"
  exit 1
else
  echo "  ✔️  Go is available!"
fi

HAVE_DOCKER=0
echo "🗲  making sure you have docker installed..."
if ! command -v docker >/dev/null 2>&1
then
    echo "  ⚠️  Warning: Docker isn't installed."
    echo "  You don't need docker if you've already got a postgres instance that you want to use for development and testing."
    echo "  Otherwise, you can go here: https://docs.docker.com/get-docker/ for install instructions."
elif ! docker info >/dev/null 2>&1
then
    echo "  ⚠️  Warning: Docker is installed but the daemon isn't running."
    echo "  You don't need docker if you've already got a postgres instance that you want to use for development and testing."
    echo "  If you're sure it's installed, try 'sudo service docker start', then re-run setup (or make pgup)."
else
    echo "  ✔️  docker is available!"
    HAVE_DOCKER=1
fi

echo "🗲  creating 'nhmlg_datavol' folder in your home directory..."
mkdir -p "$HOME/nhmlg_datavol/migrations"

if [ ! -f ./.env ]; then
  echo "🗲  creating standard '.env' file for a quick start..."
  cat << EOF > ./.env
PG_PORT=5432
DB_DATAPATH=$HOME/nhmlg_datavol/nhmlg
PG_URL=postgres://postgres:localpass@localhost:5432/localdb
MIGPATH=$HOME/nhmlg_datavol/migrations
EOF
else
  echo "  ✔️  .env already exists, leaving it as-is."
fi

echo "🗲  installing gotestsum for readable test reports..."
go install gotest.tools/gotestsum@latest

echo "🗲  building nhmlg..."
make build

if [ "$HAVE_DOCKER" -eq 1 ]; then
  echo "🗲  starting local postgres..."
  make pgup
else
  echo "  ⚠️  Skipping local postgres. Start docker and run 'make pgup', or point PG_URL in .env at your own instance."
fi

echo "✨  Done! Binary is at ./build/nhmlg"
echo "    Try: ./build/nhmlg --help"
