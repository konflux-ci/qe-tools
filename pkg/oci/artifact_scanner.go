package oci

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
)

// NewArtifactScanner creates a new instance of ArtifactScanner.
// It requires a valid ScannerConfig.
func NewArtifactScanner(cfg ScannerConfig) (*ArtifactScanner, error) {
	return &ArtifactScanner{
		config:       cfg,
		FilesPathMap: FilesPathMap{},
	}, nil
}

// Run method processes pulls the OCI artifact and stores required files (their path and content)
// in the map (ArtifactsFilesPathMap).
func (as *ArtifactScanner) Run() error {
	artifactDirPath, err := os.MkdirTemp("", "artifact-content")
	if err != nil {
		return fmt.Errorf("failed to create temporary director for pulling OCI artifact to: %+v", err)
	}
	as.artifactDirPath = artifactDirPath

	if err := as.pullAndExtractOciArtifact(); err != nil {
		return fmt.Errorf("failed to pull OCI artifact: %+v", err)
	}

	if err := as.processExtractedFiles(); err != nil {
		return fmt.Errorf("failed to process extracted files: %+v", err)
	}

	return nil
}

func (as *ArtifactScanner) pullAndExtractOciArtifact() error {
	app := "oras"
	args := []string{"pull", as.config.OciArtifactReference, "--output", as.artifactDirPath}
	// #nosec G204
	cmd := exec.Command(app, args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// processExtractedFiles is a helper function to process extracted files.
func (as *ArtifactScanner) processExtractedFiles() error {
	err := filepath.Walk(as.artifactDirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to visit file in path %s: %+v", filePath, err)
		}
		if as.isRequiredFile(filePath) {
			if err := as.initArtifactsFilesPathMap(info.Name(), filePath); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// initArtifactsFilesPathMap is  function to initialise/update the ArtifactsFilesPathMap with content
// of a file with given file path and file name
func (as *ArtifactScanner) initArtifactsFilesPathMap(fileName, filePath string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	artifact := Artifact{Content: string(fileData), Filename: fileName}

	as.FilesPathMap[FilePath(filePath)] = artifact

	return nil
}

// isRequiredFile is a helper function to check if a file with given 'ArtifactFullPath',
// matches the file-name filter(s) defined within ScannerConfig struct
func (as *ArtifactScanner) isRequiredFile(filePath string) bool {
	return slices.ContainsFunc(as.config.FileNameFilter, func(s string) bool {
		re := regexp.MustCompile(s)
		return re.MatchString(filePath)
	})
}
