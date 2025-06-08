package oci

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// Constants for configurable settings
const (
	blobTimeout = 2 * time.Minute
)

// ProcessTag processes individual tags from a given repository
func (c *Controller) ProcessTag(repo, tag, creationDate string) error {
	ctx, cancel := context.WithTimeout(context.Background(), blobTimeout)
	defer cancel()

	repoRemote, err := c.setupRemoteRepository(repo)
	if err != nil {
		return err
	}

	if err := c.copyTagManifest(ctx, repoRemote, tag, c.Store); err != nil {
		return err
	}

	outputDir := c.createOutputDirectory(repo, creationDate, tag)
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	return c.ociCopy(repoRemote, tag, outputDir)
}

func (c *Controller) ociCopy(repo *remote.Repository, tag string, outputDir string) error {
	dst, err := file.New(outputDir)
	if err != nil {
		log.Fatalf("Failed to create file store: %v", err)
	}

	defer dst.Close()

	_, err = oras.Copy(context.Background(), repo, tag, dst, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to pull artifact using oras.Copy: %v", err)
	}

	return err
}

// Sets up the remote repository for the given repo name
func (c *Controller) setupRemoteRepository(repo string) (*remote.Repository, error) {
	repoRemote, err := remote.NewRepository("quay.io/" + repo)
	if err != nil {
		return nil, fmt.Errorf("failed to set up remote repository %s: %w", repo, err)
	}

	credStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store: %w", err)
	}

	repoRemote.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore),
	}

	return repoRemote, nil
}

// Copies the tag manifest from the remote repository to the local OCI store
func (c *Controller) copyTagManifest(ctx context.Context, repoRemote *remote.Repository, tag string, store *oci.Store) error {
	if _, err := oras.Copy(ctx, repoRemote, tag, store, tag, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("failed to copy manifest for tag %s: %w", tag, err)
	}
	return nil
}

// Creates the output directory for the blobs
func (c *Controller) createOutputDirectory(repo, creationDate, tag string) string {
	parsedDate, _ := time.Parse(time.RFC1123, creationDate)
	return filepath.Join(c.OutputDir, repo, parsedDate.Format("2006-01-02"), tag)
}
