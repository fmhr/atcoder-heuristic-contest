package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func convertPathToName(path string) string {
	return strings.ReplaceAll(path, string(os.PathSeparator), "-")
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	root := filepath.Join(os.Getenv("HOME"), "ahc")
	relPath, err := filepath.Rel(root, cwd)
	if err != nil {
		fmt.Println("Error getting relative path:", err)
		return
	}

	converted := convertPathToName(relPath)
	fmt.Println(converted)
}
