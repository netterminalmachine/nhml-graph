# Contributing

Welcome to the **nhml-graph** project!

### Pre-requisites

- Please use a linux environment for development (Linux, Windows WSL2, Mac all okay)
- Make sure you have the following installed and working:
  - Docker (including docker compose)
  - Git
  - Go

### Dev Setup

To set up your dev environment for working with this repo, run `make setup`.

Here are the list of commands you can use in your local sandbox:

| command        | description                                                                                            |
| -------------- | ------------------------------------------------------------------------------------------------------ |
| `make setup`   | Set up your local sandbox for working on nhml-graph. Probably only need to do this once!             |
| `make pgup`    | Start a containerized postgres instance (it will run on the default `:5432` port (but see note below). |
| `make pgdown`  | Bring down the database                                                                                |
| `make test`    | tests the application and generates a test report                                                      |
| `make testcov` | generates a coverage report                                                                            |
| `make clean`   | deletes everything in the `build/` directory.                                                          |
| `make ls`      | lists all the migration files in the migrations folder.                                                |

:pencil: **Notes**:

- If you already have a db on :5432, you can manually edit your generated `.env` file and change all refs from **5432** to some other unused port number.
- The postgres db is `ngraph` by default, for simplicity, and the user name and password are just `postgres` and `localpass`. It doesn't matter as this is just your local, fairly throwawawy, postgres instance. You can change these values by altering the `.env` if you _really_ need to.

## Exercising the app

It's handy to have an alias to save on keystrokes, so you might want run the following:

```sh
alias nhmlg='./build/nhmlg'
```

:point_up: unfortunately you have to run this command interactively; that's just how bash works. I can't automate it for you.
After this, you can just type `nhmlg` whenever you need to invoke the app. You can also check out the main [readme](./README.md).

# Contributing/Submitting changes

Please check [the list of issues](https://github.com/netterminalmachine/nhml-graph/issues) to see if you can help with something!
Alternatively, you can **create a new issue** related to what you want to work on. Then:

- Check out a new branch with the name of the issue number. eg:
  - `$ git checkout -b fix-123/my-short-description` for fixing a bug, or:
  - `$ git checkout -b feature-123/my-short-description` for an enhancement
- Make your changes
  - If it's a code change, make sure to provide the appropriate tests (and test to make sure they pass!)
  - If it's just a documentation change, just give it a quick read-through to make sure it's easy to understand!
- Commit your changes (Always provide a git message that explains what you've done).
- Make a pull request

We'll take a look and if your PR is clear to understand, and has a small / manageable scope that reflects the issue, your PR should be approved! Remember, small PRs are better than big PRs ;-).
