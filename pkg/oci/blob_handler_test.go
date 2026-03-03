package oci

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetGzFilesFromDir(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        map[string]string // path -> content
		expectedFileCount int
		expectedError     bool
	}{
		{
			name: "Directory with .gz files",
			setupFiles: map[string]string{
				"file1.gz":  "content1",
				"file2.gz":  "content2",
				"file3.txt": "not a gz file",
			},
			expectedFileCount: 2,
		},
		{
			name: "Nested directories with .gz files",
			setupFiles: map[string]string{
				"file1.gz":               "content1",
				"subdir/file2.gz":        "content2",
				"subdir/nested/file3.gz": "content3",
				"other.txt":              "not gz",
			},
			expectedFileCount: 3,
		},
		{
			name: "No .gz files",
			setupFiles: map[string]string{
				"file1.txt": "content1",
				"file2.log": "content2",
			},
			expectedFileCount: 0,
		},
		{
			name: "Mixed file types",
			setupFiles: map[string]string{
				"archive.tar.gz": "compressed tar",
				"data.json.gz":   "compressed json",
				"readme.md":      "markdown",
				"image.png":      "image data",
			},
			expectedFileCount: 2,
		},
		{
			name:              "Empty directory",
			setupFiles:        map[string]string{},
			expectedFileCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Setup test files
			for filePath, content := range tt.setupFiles {
				fullPath := filepath.Join(tmpDir, filePath)
				dir := filepath.Dir(fullPath)

				// Create parent directories if needed
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("failed to create directory %s: %v", dir, err)
				}

				// Create file with content
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("failed to create file %s: %v", fullPath, err)
				}
			}

			// Create controller and run test
			controller := &Controller{
				OutputDir: tmpDir,
			}

			gzFiles, err := controller.GetGzFilesFromDir(tmpDir)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(gzFiles) != tt.expectedFileCount {
				t.Errorf("expected %d .gz files, got %d", tt.expectedFileCount, len(gzFiles))
			}

			// Verify all returned files end with .gz
			for _, gzFile := range gzFiles {
				if !strings.HasSuffix(gzFile.FilePath, ".gz") {
					t.Errorf("file path %q does not end with .gz", gzFile.FilePath)
				}

				// Verify DirPath is set correctly
				expectedDir := filepath.Dir(gzFile.FilePath)
				if gzFile.DirPath != expectedDir {
					t.Errorf("expected DirPath %q, got %q", expectedDir, gzFile.DirPath)
				}

				// Verify file exists
				if _, err := os.Stat(gzFile.FilePath); os.IsNotExist(err) {
					t.Errorf("file %q does not exist", gzFile.FilePath)
				}
			}
		})
	}
}

func TestGetGzFilesFromDir_NonExistentDirectory(t *testing.T) {
	controller := &Controller{}

	_, err := controller.GetGzFilesFromDir("/non/existent/directory/path")

	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestExtractGzFile(t *testing.T) {
	tests := []struct {
		name              string
		gzContent         string
		gzipName          string // Name to set in gzip header
		expectedContent   string
		expectedFileName  string
		shouldCreateEmpty bool
		expectError       bool
		skipExtraction    bool
	}{
		{
			name:             "Valid .gz file",
			gzContent:        "This is test content for extraction",
			gzipName:         "test-file.txt",
			expectedContent:  "This is test content for extraction",
			expectedFileName: "test-file.txt",
		},
		{
			name:             "Valid .gz file without gzip Name",
			gzContent:        "Content without name in header",
			gzipName:         "",
			expectedContent:  "Content without name in header",
			expectedFileName: "test.txt", // Will use filename without .gz
		},
		{
			name:             "Valid .gz file with different content",
			gzContent:        "Different test content\nwith multiple lines\n",
			gzipName:         "multi-line.log",
			expectedContent:  "Different test content\nwith multiple lines\n",
			expectedFileName: "multi-line.log",
		},
		{
			name:              "Empty .gz file (size 0)",
			shouldCreateEmpty: true,
			skipExtraction:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories
			tmpDir := t.TempDir()
			destDir := filepath.Join(tmpDir, "extracted")
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				t.Fatalf("failed to create destination directory: %v", err)
			}

			// Create gzipped file
			gzFilePath := filepath.Join(tmpDir, "test.txt.gz")

			if tt.shouldCreateEmpty {
				// Create an empty file
				if err := os.WriteFile(gzFilePath, []byte{}, 0o644); err != nil {
					t.Fatalf("failed to create empty .gz file: %v", err)
				}
			} else {
				// Create and write gzipped content
				gzFile, err := os.Create(gzFilePath)
				if err != nil {
					t.Fatalf("failed to create .gz file: %v", err)
				}
				defer gzFile.Close()

				gzWriter := gzip.NewWriter(gzFile)
				if tt.gzipName != "" {
					gzWriter.Name = tt.gzipName
				}
				if _, err := gzWriter.Write([]byte(tt.gzContent)); err != nil {
					t.Fatalf("failed to write to gzip writer: %v", err)
				}
				if err := gzWriter.Close(); err != nil {
					t.Fatalf("failed to close gzip writer: %v", err)
				}
			}

			// Create controller and extract
			controller := &Controller{
				OutputDir: tmpDir,
			}

			err := controller.ExtractGzFile(gzFilePath, destDir)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.skipExtraction {
				// For empty files, extraction should be skipped
				// Verify no files were created in destDir
				entries, err := os.ReadDir(destDir)
				if err != nil {
					t.Fatalf("failed to read destination directory: %v", err)
				}
				if len(entries) > 0 {
					t.Errorf("expected no files in destination directory for empty .gz file, found %d", len(entries))
				}
				return
			}

			// Verify extracted file exists
			extractedFilePath := filepath.Join(destDir, tt.expectedFileName)
			if _, err := os.Stat(extractedFilePath); os.IsNotExist(err) {
				t.Fatalf("extracted file %q does not exist", extractedFilePath)
			}

			// Verify extracted content
			extractedContent, err := os.ReadFile(extractedFilePath)
			if err != nil {
				t.Fatalf("failed to read extracted file: %v", err)
			}

			if string(extractedContent) != tt.expectedContent {
				t.Errorf("expected content %q, got %q", tt.expectedContent, string(extractedContent))
			}
		})
	}
}

func TestExtractGzFile_InvalidGzFile(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}

	// Create a file that's not actually gzipped
	invalidGzPath := filepath.Join(tmpDir, "invalid.gz")
	if err := os.WriteFile(invalidGzPath, []byte("not gzipped content"), 0o644); err != nil {
		t.Fatalf("failed to create invalid .gz file: %v", err)
	}

	controller := &Controller{
		OutputDir: tmpDir,
	}

	err := controller.ExtractGzFile(invalidGzPath, destDir)

	if err == nil {
		t.Error("expected error for invalid .gz file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create gzip reader") {
		t.Errorf("expected error about gzip reader, got: %v", err)
	}
}

func TestExtractGzFile_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	controller := &Controller{
		OutputDir: tmpDir,
	}

	err := controller.ExtractGzFile("/non/existent/file.gz", tmpDir)

	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open .gz file") {
		t.Errorf("expected error about opening file, got: %v", err)
	}
}
