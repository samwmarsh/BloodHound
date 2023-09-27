package workgen

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "workgen"
	Usage = "Run all code generation in workspace"
)

type cmd struct{}

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
	} else if err := workspace.WorkspaceGenerate(modPaths); err != nil {
		return fmt.Errorf("could not generate code in modules: %w", err)
	} else {
		return nil
	}
}

func Create() (cmd, error) {
	var (
		command    = cmd{}
		workgenCmd = flag.NewFlagSet(Name, flag.ExitOnError)
	)

	workgenCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		workgenCmd.PrintDefaults()
	}

	if err := workgenCmd.Parse(os.Args[2:]); err != nil {
		workgenCmd.Usage()
		return command, fmt.Errorf("failed to parse workgen command: %w", err)
	} else {
		return command, nil
	}
}
