This tool combines files from a Local repository into a single output file.

## Usage

```
go run combine_files.go -dir=<repository_path> [options]
```

## Options

- `-dir string`
  - Path to the GitHub repository (required)
- `-out string`
  - Path to the output file (default: `combined_output.txt`)
- `-ext string`
  - Target file extensions (comma-separated, e.g., `py,js,md`)
- `-exclude string`
  - Exclude directories (comma-separated, default: `.git,node_modules,venv,dist,build`)
- `-help`
  - Show this help message

## Example

```
go run combine_files.go -dir=./my-repo -out=combined.txt -ext=py,js,md
```

## Functionality

The tool recursively walks through the specified repository, filters files based on the provided extensions and excluded directories, and combines the content of the remaining files into a single output file.

### Directory Structure

The output file starts with a section that lists the directory structure of the repository.

### File Combination

Each file's content is added to the output file, preceded by a header indicating the file path.

## Flags

- **-dir**: Specifies the path to the repository. This flag is required.
- **-out**: Specifies the path to the output file. If not provided, the default value is `combined_output.txt`.
- **-ext**: Specifies the target file extensions. This is a comma-separated list of extensions.
- **-exclude**: Specifies the directories to exclude. This is a comma-separated list of directory names.
- **-help**: Displays the help message.
