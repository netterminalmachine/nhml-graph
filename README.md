# nhml-graph

This project is booted from a fork of my old [nano-migrate](https://github.com/lululeon/nano-migrate) project, which was just a vehicle for learning go.

An env-aware, forward-only, postgres migrator for solo devs and tiny teams.

- :computer: A simple CLI tool
- :zap: recognizes `.env` files and environment variables for easy substitution into sql migration files
- :arrow_up: no messing around with down migrations: roll with the punches and roll forward.

## Installation

(todo)

## Quickstart for developers

read the [CONTRIBUTING](./CONTRIBUTING.md) guide. TLDR: `make setup`, alias `nhmlg` to `./build/nhmlg`, and you're good to go.

## General Usage
Run the following command to initialise the migrator

```sh
nhmlg init
```
You should see a message like 'Migrations table ready.'

To create a brand new migration, simply use `nhmlg create` followed by a description of the schema change. e.g.:

```sh
nhmlg create initial foo table
```

:point_up: You can also just use 'c' as an alias for 'create'. 

This will create a new file in your migrations folder, named `nnnn-initial-foo-table.sql`, or similar. It will be empty, because nhml-graph doesn't opine on how you write your migrations and it doesn't support down migrations anyway, so really... enter whatever sql you'd like to run! :-)

:warning: **Do not use transactions**, as your migration will be wrapped in a postgres transaction.

To run all pending migrations that have not yet been applied to your database:

```sh
nhmlg migrate
```

That's it!

---

## Contributing

Here's our tiny [contributing](./CONTRIBUTING.md) guide. We especially welcome contributors who are learning go as well as people who want the experience of contributing to a project! :-)
