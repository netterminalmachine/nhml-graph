#! /bin/bash

if [ ! -f ./.env ]; then
  echo ".env file not found. you must have one to proceed! bye 👋"
  exit 1
fi

export $(grep -v '^\#' .env | grep 'DB_DATAPATH')

echo "🚨💥The recycle command will nuke your postgres volume at $DB_DATAPATH"
echo "This means you will lose ABSOLUTELY ALL THE DATA in your local database."
echo 

# -n 1 : accept only one character. -r : raw input (no backslashes/escapes). -s : silent (don't echo the input).
read -p "Proceed? y/n" -n 1 -r -s
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]
then
  echo "Deleting postgres volume $DB_DATAPATH"
  sudo rm -rf $DB_DATAPATH

  echo "recreating data volume..."
  mkdir $DB_DATAPATH

  ls $DB_DATAPATH

  echo "✅ Done."
else
  echo "Good call. ✌️"
fi
