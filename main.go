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

	outFile, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error: Failed to create the output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, "# Directory structure")
	fmt.Fprintln(outFile, "```")

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
	err = filepath.Walk(*repoPath, func(path string, info os.FileInfo, err error) error {
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
	for _, path := range filePaths {
		fmt.Fprintln(outFile, path)
	}

	fmt.Fprintln(outFile, "```")
	fmt.Fprintln(outFile, "")

	fmt.Println("Collecting files from the repository...")
	fileCount, err := combineFiles(*repoPath, outFile, extFilter, excludeDirList)
	if err != nil {
		fmt.Printf("Error: Failed to combine files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done! Combined %d files into '%s'\n", fileCount, *outputFile)
}

func printUsage() {
	fmt.Println("A tool to combine files from a GitHub repository into one")
	fmt.Println("\nUsage:")
	fmt.Println("  go run combine_files.go -repo=<repository_path> [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -dir string     Path to the GitHub repository(required)")
	fmt.Println("  -out string   Path to the output file(default: \"combined_output.txt\")")
	fmt.Println("  -ext string       Target file extensions(comma-separated, e.g. 'py,js,md')")
	fmt.Println("  -exclude string   Exclude directories(comma-separated, default: \".git,node_modules,venv,dist,build\")")
	fmt.Println("  -help             Show this help")
	fmt.Println("\nExample:")
	fmt.Println("  go run combine_files.go -dir=./my-repo -out=combined.txt -ext=py,js,md")
}

func combineFiles(repoPath string, outFile io.Writer, extFilter []string, excludeDirs []string) (int, error) {
	fileCount := 0

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, excludeDir := range excludeDirs {
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

		if isBinaryFile(path) {
			return nil
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		fmt.Printf("Processing: %s\n", relPath)

		fmt.Fprintln(outFile, "----------------------------------------")
		fmt.Fprintf(outFile, "# File: %s\n", relPath)
		fmt.Fprintln(outFile, "----------------------------------------")
		fmt.Fprintln(outFile, "")

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = outFile.Write(content)
		if err != nil {
			return err
		}

		fmt.Fprintln(outFile, "")
		fmt.Fprintln(outFile, "")

		fileCount++
		return nil
	})

	return fileCount, err
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
