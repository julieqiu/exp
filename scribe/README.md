# Scribe

Scribe is a tool for experimenting with Google Cloud client libraries
documentation. It runs a local web server to help you browse and explore the
existing library reference documentation that has been collected.

## Running the Documentation Server

To start the local server and browse the documentation, run the `docs` command:

```sh
go run ./cmd/scribe docs
```

By default, the server runs on port `8080`. You can open your web browser and
navigate to `http://localhost:8080` to view the documentation.

If you need to use a different port, you can specify it with the `--port` flag:

```sh
go run ./cmd/scribe docs --port=8888
```

## Refreshing Documentation Data

The documentation data is pre-scraped and included in the `data/` directory. If
you need to refresh this data, you can use the `scrape` command.

To scrape the documentation for a specific language:

```sh
go run ./cmd/scribe scrape <language>
```

For example, to scrape for Python libraries:
```sh
go run ./cmd/scribe scrape python
```

To update the data for all supported languages, use the `--all` flag:

```sh
go run ./cmd/scribe scrape --all
```

