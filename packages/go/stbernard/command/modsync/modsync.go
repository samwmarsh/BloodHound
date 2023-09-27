package modsync

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "modsync"
	Usage = "Sync all modules in current workspace"
)

type cmd struct {
	Verbose bool
}

func (s cmd) Usage() string {
	return Usage
}

func (s cmd) Name() string {
	return Name
}

func (s cmd) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("could not find workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse module absolute paths: %w", err)
	} else if err := workspace.DownloadModules(modPaths); err != nil {
		return fmt.Errorf("could not download modules: %w", err)
	} else if err := workspace.SyncWorkspace(cwd); err != nil {
		return fmt.Errorf("could not sync workspace: %w", err)
	} else {
		return nil
	}
}

func Create() (cmd, error) {
	var (
		command          = cmd{}
		verboseEnv       = os.Getenv(environment.Verbose.Env())
		modsyncCmd       = flag.NewFlagSet(Name, flag.ExitOnError)
		verboseBool, err = strconv.ParseBool(verboseEnv)
	)

	if err != nil {
		return command, fmt.Errorf("failed to parse environment variable (%s=%s) as a boolean: %w", "verbose", verboseEnv, err)
	}

	modsyncCmd.BoolVar(&command.Verbose, "v", verboseBool, "Print verbose logs")
	modsyncCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		modsyncCmd.PrintDefaults()
	}

	if err := modsyncCmd.Parse(os.Args[2:]); err != nil {
		modsyncCmd.Usage()
		return command, fmt.Errorf("failed to parse modsync command: %w", err)
	} else {
		return command, nil
	}
}
