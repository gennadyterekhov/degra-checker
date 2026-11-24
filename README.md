# degra-checker

Checks the direction of imports between layers.

## Configuration

Create `import_layers.json` in the project root.

```json
{
  "layers": ["Domain", "Application"],
  "debug": false,
  "ignoreTests": false
}
```

The analyzer searches upward from the working directory for the nearest configuration.
It looks for `import_layers.json`, `import_layers.yaml`, and `import_layers.yml`,
in that order within each directory. Use `-c` or `--config` to select a different JSON or YAML file.
Its path is relative to the working directory. 
The project root remains the directory containing the nearest default configuration,
or the working directory when there is no default configuration.

## Usage 

1. ast-go to get ast
2. degra to get imports graph
3. degra-checker to check import layers
