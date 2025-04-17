package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	repoPath := flag.String("dir", "", "Path to the repository(required)")
	outputFile := flag.String("out", "combined_output.txt", "Path to the output file")
	extensions := flag.String("ext", "", "Target file extensions(comma-separated, e.g. 'py,js,md')")
	excludeDirs := flag.String("exclude", ".git,node_modules,venv,dist,build", "Exclude directories(comma-separated)")
	maxFileSize := flag.Int64("maxsize", 10*1024*1024, "Maximum file size in bytes before splitting output to a new file")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help || *repoPath == "" {
		printUsage()
		os.Exit(0)
	}

	if _, err := os.Stat(*repoPath); os.IsNotExist(err) {
		fmt.Printf("Error: The specified repository path '%s' does not exist\n", *repoPath)
		os.Exit(1)
	}

	excludeDirList := strings.Split(*excludeDirs, ",")

	// 拡張子フィルタの準備
	var extFilter []string
	if *extensions != "" {
		extFilter = strings.Split(*extensions, ",")
		for i, ext := range extFilter {
			extFilter[i] = "." + strings.TrimPrefix(ext, ".")
		}
	}

	var filePaths []string
	err := filepath.Walk(*repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, excludeDir := range excludeDirList {
				if info.Name() == excludeDir || strings.Contains(path, "/"+excludeDir+"/") {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		if len(extFilter) > 0 {
			ext := filepath.Ext(path)
			found := false
			for _, validExt := range extFilter {
				if ext == validExt {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		relPath, err := filepath.Rel(*repoPath, path)
		if err != nil {
			return err
		}

		filePaths = append(filePaths, relPath)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: Failed to collect files: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(filePaths)

	fmt.Println("Collecting files from the repository...")
	fileCount, outputFiles, err := combineFiles(*repoPath, *outputFile, extFilter, excludeDirList, filePaths, *maxFileSize)
	if err != nil {
		fmt.Printf("Error: Failed to combine files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done! Combined %d files into %d output files: %s\n", fileCount, len(outputFiles), strings.Join(outputFiles, ", "))
}

func printUsage() {
	fmt.Println("A tool to combine files from a Local repository into one")
	fmt.Println("\nUsage:")
	fmt.Println("  go run main.go -dir=<repository_path> [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -dir string     Path to the repository(required)")
	fmt.Println("  -out string     Path to the output file(default: \"combined_output.txt\")")
	fmt.Println("  -ext string     Target file extensions(comma-separated, e.g. 'py,js,md')")
	fmt.Println("  -exclude string Exclude directories(comma-separated, default: \".git,node_modules,venv,dist,build\")")
	fmt.Println("  -maxsize int    Maximum file size in bytes before splitting output to a new file(default: 10MB)")
	fmt.Println("  -help           Show this help")
	fmt.Println("\nExample:")
	fmt.Println("  go run main.go -dir=./my-repo -out=combined.txt -ext=py,js,md -maxsize=5242880")
	fmt.Println("  or")
	fmt.Println("  go run github.com/nakamurakzz/merge-repo@latest -dir=./my-repo -out=combined.txt -ext=py,js,md -maxsize=5242880")
	fmt.Println("")
}

func getNextFileName(baseFileName string, sequenceNum int) string {
	ext := filepath.Ext(baseFileName)
	baseName := strings.TrimSuffix(baseFileName, ext)
	return fmt.Sprintf("%s_%d%s", baseName, sequenceNum, ext)
}

func createNewOutputFile(fileName string) (*os.File, error) {
	outFile, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create the output file: %v", err)
	}
	fmt.Println("Created new output file:", fileName)

	// Write directory structure header to new file
	fmt.Fprintln(outFile, "# Directory structure")
	fmt.Fprintln(outFile, "```")
	return outFile, nil
}

func combineFiles(repoPath, outputFileName string, extFilter, excludeDirs, filePaths []string, maxFileSize int64) (int, []string, error) {
	fileCount := 0
	currentSize := int64(0)
	sequenceNum := 1
	outputFiles := []string{outputFileName}

	// Create first output file
	outFile, err := createNewOutputFile(outputFileName)
	if err != nil {
		return 0, nil, err
	}
	defer outFile.Close()

	// Write file paths to first file
	for _, path := range filePaths {
		fmt.Fprintln(outFile, path)
	}

	fmt.Fprintln(outFile, "```")
	fmt.Fprintln(outFile, "")

	// Update current size with directory structure
	fileInfo, err := outFile.Stat()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get file stats: %v", err)
	}
	currentSize = fileInfo.Size()

	for _, relPath := range filePaths {
		fullPath := filepath.Join(repoPath, relPath)

		if isBinaryFile(fullPath) {
			continue
		}

		fmt.Printf("Processing: %s\n", relPath)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fileCount, outputFiles, err
		}

		// Calculate content size with file header
		headerText := fmt.Sprintf("\n----------------------------------------\n# File: %s\n----------------------------------------\n\n", relPath)
		footerText := "\n\n"
		entrySize := int64(len(content) + len(headerText) + len(footerText))

		// Check if adding this file would exceed maxFileSize
		if currentSize+entrySize > maxFileSize && fileCount > 0 {
			// Close current file
			outFile.Close()

			// Create new output file with sequential name
			nextFileName := getNextFileName(outputFileName, sequenceNum)
			outFile, err = createNewOutputFile(nextFileName)
			if err != nil {
				return fileCount, outputFiles, err
			}
			defer outFile.Close()

			// Add new filename to the list
			outputFiles = append(outputFiles, nextFileName)
			sequenceNum++
			currentSize = 0

			// Write directory structure closing tag
			fmt.Fprintln(outFile, "```")
			fmt.Fprintln(outFile, "")

			// Update current size
			fileInfo, err := outFile.Stat()
			if err != nil {
				return fileCount, outputFiles, fmt.Errorf("failed to get file stats: %v", err)
			}
			currentSize = fileInfo.Size()
		}

		// Write file header
		fmt.Fprint(outFile, headerText)

		// Write file content
		_, err = outFile.Write(content)
		if err != nil {
			return fileCount, outputFiles, err
		}

		// Write footer
		fmt.Fprint(outFile, footerText)

		fileCount++
		currentSize += entrySize
	}

	return fileCount, outputFiles, nil
}

func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}
