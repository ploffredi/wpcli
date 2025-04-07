# WPStore CLI

A command line interface for managing WordPress plugins from the wpstore repository.

## Installation

```bash
go install github.com/ploffredi/wpcli@latest
```

## Usage

The CLI provides the following commands:

### List all plugins

```bash
wpcli list
```

This command will display a list of all available plugins from the wpstore repository.

### Get plugin information

```bash
wpcli info [plugin-name]
```

This command will display detailed information about a specific plugin.

## Development

To build the CLI from source:

```bash
git clone https://github.com/ploffredi/wpcli.git
cd wpcli
go build
```

## License

MIT
