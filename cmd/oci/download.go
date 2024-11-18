package download

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konflux-ci/qe-tools/pkg/oci"
	"github.com/spf13/cobra"
)

// downloadOptions holds the configuration options for the download command.
// It contains flags that determine how artifacts are downloaded from OCI storage.
type downloadOptions struct {
	// repo specifies a single OCI repository and tag from which to download artifacts.
	// This is used when the user wants to download from a specific repository.
	repo string

	// repos is a slice of strings representing multiple OCI repositories.
	// This allows the user to specify a list of repositories to download from.
	// This option should be used in conjunction with the `--since` flag to filter artifacts.
	repos []string

	// since specifies a time range for downloading the latest artifacts.
	// It accepts durations in various formats (e.g., "4h", "2d").
	// This flag is required when downloading from multiple repositories specified in the `repos` field.
	since string

	// ociCache specifies the directory where OCI artifacts will be cached.
	// If not provided, a default directory will be created at $HOME/.config/qe-tools/cache.
	ociCache string

	// artifactsOutput specifies the output path for downloaded artifacts.
	// This field is mandatory and specifies where the downloaded artifacts should be stored.
	artifactsOutput string

	// noCache determines whether to remove the OCI cache after downloading artifacts.
	// If true, the cache will be deleted after the command execution completes, regardless of success or failure.
	noCache bool

	// uncompressGzFiles will extract all the .gz files from the oci artifacts
	uncompressGzFiles bool
}

var opts = &downloadOptions{}

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download artifacts from OCI storage",
	Long: `Download artifacts from OCI storage based on the specified repository or repositories.

Examples:
  - Download from a single repository:
      qe-tools download --repo quay.io/test/test:1.0 --artifacts-output /path/to/output

  - Download from multiple repositories with a time range:
      qe-tools download --repos quay.io/repo1 quay.io/repo2 --since 4h --artifacts-output /path/to/output
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

		// Validation: Fail if both 'repo' and 'repos' are provided
		if opts.repo != "" && len(opts.repos) > 0 {
			return fmt.Errorf("you cannot use both --repo and --repos at the same time")
		}

		// Validation: Fail if 'since' is provided without 'repos', or 'repos' is provided without 'since'
		if opts.since != "" && len(opts.repos) == 0 {
			return fmt.Errorf("the --since flag requires the --repos flag")
		}
		if len(opts.repos) > 0 && opts.since == "" {
			return fmt.Errorf("the --repos flag requires the --since flag")
		}

		// If neither 'repo' nor 'repos' is provided, show command-specific help
		if opts.repo == "" && len(opts.repos) == 0 {
			if err := cmd.Help(); err != nil {
				return fmt.Errorf("failed to display help: %w", err)
			}
			return fmt.Errorf("either --repo or --repos must be specified")
		}

		// Check if the mandatory artifactsOutput flag is provided
		if opts.artifactsOutput == "" {
			return fmt.Errorf("the --artifacts-output flag is mandatory")
		}

		// Set the default OCI cache directory if not specified
		if opts.ociCache == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %v", err)
			}
			opts.ociCache = filepath.Join(homeDir, ".config", "qe-tools", "cache")
		}

		// Create the cache directory if it doesn't exist
		if err := os.MkdirAll(opts.ociCache, 0o750); err != nil {
			return fmt.Errorf("could not create cache directory: %v", err)
		}

		// Use defer to ensure cache removal at the end
		defer func() {
			if opts.noCache {
				if err := os.RemoveAll(opts.ociCache); err != nil {
					log.Printf("Warning: could not remove cache directory: %v\n", err)
				}
			}
		}()

		ociController, err := oci.NewController(opts.artifactsOutput, opts.ociCache)
		if err != nil {
			return fmt.Errorf("failed to create OCI controller with artifactsOutput: '%s' and ociCache: '%s': %v", opts.artifactsOutput, opts.ociCache, err)
		}

		// If repo is specified, call helper function to download from a single repository
		if opts.repo != "" {
			repo, tag, err := parseRepoAndTag(opts.repo)
			if err != nil {
				return err
			}

			// Call ProcessTag to get details of the tag (implement as needed)
			if err := ociController.ProcessTag(repo, tag, time.Now().Format(time.RFC1123)); err != nil {
				return fmt.Errorf("failed to fetch tag: %v", err)
			}
		}

		// Handle time-based downloads
		if opts.since != "" {
			duration, err := parseDuration(opts.since)
			if err != nil {
				return fmt.Errorf("invalid time format for --since: %v", err)
			}
			log.Printf("Downloading latest artifacts from the last %s\n", duration)
		}

		// If repos is specified, simulate a download from multiple repositories
		if len(opts.repos) > 0 {
			var processRepos []string
			var allErrors []error
			for _, repo := range opts.repos {
				if !strings.HasPrefix(repo, "quay.io/") {
					return fmt.Errorf("the repository must start with 'quay.io/'")
				}

				// Remove 'quay.io/' prefix
				repo = strings.TrimPrefix(repo, "quay.io/")
				processRepos = append(processRepos, repo)
			}

			if opts.since != "" {
				duration, err := parseDuration(opts.since)
				if err != nil {
					return fmt.Errorf("invalid time format for --since: %v", err)
				}
				errors := ociController.ProcessRepositories(processRepos, duration)
				allErrors = append(allErrors, errors...)
			}

			if len(allErrors) > 0 {
				log.Println("Errors encountered during processing:")
				for _, err := range allErrors {
					log.Printf(" - %v\n", err)
				}
			}
		}

		if opts.uncompressGzFiles {
			gzFilesFromOciArtifacts, err := ociController.GetGzFilesFromDir(opts.artifactsOutput)
			if err != nil {
				return fmt.Errorf("failed to extract gz files from artifacts folder: %w", err)
			}

			for _, file := range gzFilesFromOciArtifacts {
				err := ociController.ExtractGzFile(file.FilePath, file.DirPath)
				if err != nil {
					log.Printf("warn: file %s was not extracted successfully: %s", file.FilePath, err.Error())
				}

				if err := os.Remove(file.FilePath); err != nil {
					log.Printf("failed to remove gz file - %v\n", err)
				}
			}
		}

		return nil
	},
}

// parseRepoAndTag extracts the repository and tag from the given repo flag.
func parseRepoAndTag(repoFlag string) (string, string, error) {
	// Ensure the repoFlag starts with 'quay.io/'
	if !strings.HasPrefix(repoFlag, "quay.io/") {
		return "", "", fmt.Errorf("the repository must start with 'quay.io/'")
	}

	// Remove 'quay.io/' prefix and split the repo and tag using the ':' character
	repoFlag = strings.TrimPrefix(repoFlag, "quay.io/")
	parts := strings.SplitN(repoFlag, ":", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("tag is missing in the repo flag")
	}

	return parts[0], parts[1], nil
}

// parseDuration handles the custom duration format
func parseDuration(since string) (time.Duration, error) {
	if len(since) > 1 && since[len(since)-1] == 'd' {
		days := since[:len(since)-1]
		hours, err := time.ParseDuration(days + "h")
		if err != nil {
			return 0, err
		}
		return hours * 24, nil
	}

	// Parse the duration normally for other time units
	duration, err := time.ParseDuration(since)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

// Init initializes the download command and its flags
func Init() *cobra.Command {
	downloadCmd.Flags().StringVar(&opts.repo, "repo", "", "OCI repository and tag to download (e.g., quay.io/test/test:1.0)")
	downloadCmd.Flags().StringSliceVar(&opts.repos, "repos", nil, "Set of OCI repositories to download from")
	downloadCmd.Flags().StringVar(&opts.since, "since", "", "Time range to download the latest artifacts (e.g., 4h, 10m, 2d)")
	downloadCmd.Flags().StringVar(&opts.ociCache, "oci-cache", "", "Directory where OCI artifacts will be cached (default: $HOME/.config/qe-tools/cache)")
	downloadCmd.Flags().StringVar(&opts.artifactsOutput, "artifacts-output", "", "Mandatory path to store downloaded artifacts")
	downloadCmd.Flags().BoolVar(&opts.noCache, "no-cache", true, "If true, removes the OCI cache after downloading artifacts")
	downloadCmd.Flags().BoolVar(&opts.uncompressGzFiles, "uncompress-gz-files", true, "If true, uncompresses all gzipped files from the OCI artifacts after download.")

	return downloadCmd
}
