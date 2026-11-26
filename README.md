# degra-checker

Checks the direction of imports between layers in a JSON graph from `degra`.

## Usage

From the `degra-checker` directory:

```sh
go build -o degra-checker ./cmd
./degra-checker --config examples/import_layers.yaml < ../degra/examples/output/imports-go.json
```

The CLI reads the import map from stdin. It also accepts the edge array produced by `degra --graph-edges`. It writes a JSON array of unique violations to stdout, sorted by source and target path.

```json
[{"from":"src/domain/model.ts","to":"src/app/service.ts","fromLayer":"src/domain","toLayer":"src/app"}]
```

Exit codes are `0` when the array is empty, `1` when it contains violations, and `2` for errors. Diagnostics go to stderr.

## Configuration

```json
{
  "layers": ["src/domain", "src/app"],
  "rootDir": "src",
  "ignoreTests": false,
  "debug": false
}
```

A layer can import itself and any earlier layer. Both files must match configured layers for an edge to be checked. Project-relative directory paths take precedence over single directory names, which match whole segments. When multiple selectors of the same kind match, configuration order decides. Python selectors can also match module segments, for example `analyzer.layer_policy`. The graph contains file paths, so PHP namespace selectors must be expressed as directory selectors.

Use `rootDir` to restrict both endpoints to one subtree. Set `ignoreTests` (or its legacy name `ignore_tests`) to exclude test files according to the language's naming conventions. `debug` logs visited files to stderr.

Use `-c` or `--config` to select a JSON or YAML file. Without a flag, the checker searches upward from the working directory for the nearest configuration, trying `import_layers.json`, `import_layers.yaml`, and `import_layers.yml` in that order in each directory. Graph paths are interpreted relative to the project root used to produce the graph.

Run tests with `go test ./...`.
