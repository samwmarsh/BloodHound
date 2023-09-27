package workspace

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/mod/modfile"
)

// FindRoot will attempt to crawl up the path until it finds a go.work file
func FindRoot() (string, error) {
	if cwd, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	} else {
		var found bool

		for !found {
			found, err = WorkFileExists(cwd)
			if err != nil {
				return cwd, fmt.Errorf("error while trying to find go.work file: %w", err)
			}

			if found {
				break
			}

			prevCwd := cwd

			// Go up a directory before retrying
			cwd = filepath.Dir(cwd)

			if cwd == prevCwd {
				return cwd, errors.New("found root path without finding go.work file")
			}
		}

		return cwd, nil
	}
}

// WorkFileExists checks if a go.work file exists in the given directory
func WorkFileExists(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not stat go.work file: %w", err)
	} else {
		return true, nil
	}
}

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("could not read go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("could not parse go.work file: %w", err)
	} else {
		var (
			modulePaths = make([]string, 0, len(workfile.Use))
			workDir     = filepath.Dir(workfilePath)
		)

		for _, use := range workfile.Use {
			modulePaths = append(modulePaths, filepath.Join(workDir, use.Path))
		}

		return modulePaths, nil
	}
}

// DownloadModules runs go mod download for all module paths passed
func DownloadModules(modPaths []string) error {
	var errs = make([]error, 0)

	for _, modPath := range modPaths {
		cmd := exec.Command("go", "mod", "download")
		cmd.Dir = modPath
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failure when running go mod download in %s: %w", modPath, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to download all modules: %w", errors.Join(errs...))
	} else {
		return nil
	}
}

// WorkspaceGenerate runs go generate ./... for all module paths passed
func WorkspaceGenerate(modPaths []string) error {
	var (
		errs = make([]error, 0)
		wg   sync.WaitGroup
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			defer wg.Done()
			cmd := exec.Command("go", "generate", "./...")
			cmd.Dir = modPath
			if err := cmd.Run(); err != nil {
				errs = append(errs, fmt.Errorf("failure when running go generate command in %s: %w", modPath, err))
			}
		}(modPath)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to generate code in all modules: %w", errors.Join(errs...))
	} else {
		return nil
	}
}

// SyncWorkspace runs go work sync in the given directory
func SyncWorkspace(cwd string) error {
	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed running go work sync: %w", err)
	} else {
		return nil
	}
}
