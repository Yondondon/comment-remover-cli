# Go Comment Remover

A CLI tool written in Go that removes comments from Go (`.go`) source files.

## Features

- Remove comments from a single Go file.
- Process all Go files in a directory.
- Recursively process subdirectories with the `--recursive` flag.
- Specify the target file or directory using the `--path` flag.

## Installation

Build the application:

```bash
go build -o cmnt-rmvr .
```

## Usage

```bash
cmnt-rmvr remove --path <path> [--recursive]
```

### Command

| Command | Description |
|---|---|
| `remove` | Remove comments from Go source files. |

### Flags

| Flag | Shorthand | Required | Description |
|---|---|---|---|
| `--path` | `-p` | Yes | Path to a Go file or directory. |
| `--recursive` | `-r` | No | Recursively process subdirectories. |

## Examples

**Remove comments from a single file:**

```bash
cmnt-rmvr remove --path ./main.go
```

**Remove comments from a directory:**

```bash
cmnt-rmvr remove --path ./src
```

**Recursively process a directory and its subdirectories:**

```bash
cmnt-rmvr remove --path ./src --recursive
```

**Use shorthand flags:**

```bash
cmnt-rmvr remove -p ./src -r
```
