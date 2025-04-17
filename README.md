# Repository File Combiner

This tool combines files from a Local repository into a single output file or multiple output files based on size constraints.

## Usage

```
go run combine_files.go -dir=<repository_path> [options]
```

## Options

- `-dir string`
  - Path to the Local repository (required)
- `-out string`
  - Base path for the output file(s) (default: `combined_output.txt`)
- `-ext string`
  - Target file extensions (comma-separated, e.g., `py,js,md`)
- `-exclude string`
  - Exclude directories (comma-separated, default: `.git,node_modules,venv,dist,build`)
- `-maxsize int`
  - Maximum file size in bytes before splitting output to a new file (default: 10MB)
- `-help`
  - Show this help message

## Example

```
go run main.go -dir=./my-repo -out=combined.txt -ext=py,js,md -maxsize=5242880
```

## Functionality

The tool recursively walks through the specified repository, filters files based on the provided extensions and excluded directories, and combines the content of the remaining files into one or more output files.

### Size-based File Splitting

When the output file reaches the specified size limit (default: 10MB), the tool automatically creates a new output file with a sequential number appended to the filename (e.g., `combined_1.txt`, `combined_2.txt`, etc.).

### Directory Structure

Each output file starts with a section that lists the directory structure of the repository.

### File Combination

Each file's content is added to the output file, preceded by a header indicating the file path.

## Flags

- **-dir**: Specifies the path to the repository. This flag is required.
- **-out**: Specifies the base path for the output file(s). If not provided, the default value is `combined_output.txt`.
- **-ext**: Specifies the target file extensions. This is a comma-separated list of extensions.
- **-exclude**: Specifies the directories to exclude. This is a comma-separated list of directory names.
- **-maxsize**: Specifies the maximum size (in bytes) for each output file before creating a new one. Default is 10MB (10*1024*1024 bytes).
- **-help**: Displays the help message.
