package oci

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GzFileInfo holds information about a gzipped file and its directory.
type GzFileInfo struct {
	// FilePath represents the full path to the gzipped file.
	FilePath string

	// DirPath represents the directory where the gzipped file are stored.
	DirPath string
}

// GetGzFilesFromDir retrieves all .gz files in the specified directory
func (c *Controller) GetGzFilesFromDir(dir string) ([]GzFileInfo, error) {
	var gzFiles []GzFileInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".gz") {
			gzFiles = append(gzFiles, GzFileInfo{
				FilePath: path,
				DirPath:  filepath.Dir(path),
			})
		}
		return nil
	})

	return gzFiles, err
}

// ExtractGzFile extracts a .gz file to the specified output directory.
// It decompresses the .gz file into its original file.
func (c *Controller) ExtractGzFile(gzFilePath, destDir string) error {
	// #nosec G304
	gzFile, err := os.Open(gzFilePath)
	if err != nil {
		return fmt.Errorf("failed to open .gz file: %w", err)
	}
	defer gzFile.Close()

	gzFileStats, err := os.Stat(gzFilePath)
	if err != nil {
		return fmt.Errorf("failed to get stat for .gz file: %w", err)
	}

	// If the file size is 0, it's considered empty. Return nil to skip extraction
	if gzFileStats.Size() == 0 {
		return nil
	}

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	outputFileName := gzReader.Name
	if outputFileName == "" {
		outputFileName = strings.TrimSuffix(filepath.Base(gzFilePath), ".gz")
	}

	outputFilePath := filepath.Join(destDir, outputFileName)
	// #nosec G304
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// #nosec G110
	_, err = io.Copy(outputFile, gzReader)
	if err != nil {
		// Check for EOF error, and treat it as normal if it occurs.
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to write decompressed data to file: %w", err)
	}

	return nil
}
