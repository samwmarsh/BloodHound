package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/command/buildcmd"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/workgen"
)

// Commander is an interface for commands, allowing commands to implement the minimum
// set of requirements to observe and run the command from above. It is used as a return
// type to allow passing a usable command to the caller after parsing and creating
// the command implementation
type Commander interface {
	Name() string
	Usage() string
	Run() error
}

var (
	ErrNoCmd      = errors.New("no command specified")
	ErrInvalidCmd = errors.New("invalid command specified")
	ErrCreateCmd  = errors.New("failed to create command")
)

// ParseCLI parses for a subcommand as the first argument to the calling binary,
// and initializes the command (if it exists). It also provides the default usage
// statement.
//
// It does not support flags of its own, each subcommand is responsible for parsing
// their flags.
func ParseCLI() (Commander, error) {
	// Generate a nice usage message
	flag.Usage = usage

	// Default usage if no arguments provided
	if len(os.Args) < 2 {
		flag.Usage()
		return nil, ErrNoCmd
	}

	switch os.Args[1] {
	case ModSync.String():
		if cmd, err := modsync.Create(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateCmd, err)
		} else {
			return cmd, nil
		}

	case EnvDump.String():
		if cmd, err := envdump.Create(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateCmd, err)
		} else {
			return cmd, nil
		}

	case WorkGen.String():
		if cmd, err := workgen.Create(); err != nil {
			return nil, fmt.Errorf("%w, %w", ErrCreateCmd, err)
		} else {
			return cmd, nil
		}

	case Build.String():
		if cmd, err := buildcmd.Create(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateCmd, err)
		} else {
			return cmd, nil
		}

	default:
		flag.Parse()
		flag.Usage()
		return nil, ErrInvalidCmd
	}
}

// usage creates a pretty usage message for our main command
func usage() {
	var longestCmdLen int

	w := flag.CommandLine.Output()
	fmt.Fprint(w, "A BloodHound Swiss Army Knife\n\nUsage:  stbernard COMMAND\n\nCommands:\n")

	for _, cmd := range Commands() {
		if len(cmd.String()) > longestCmdLen {
			longestCmdLen = len(cmd.String())
		}
	}

	for cmd, usage := range CommandsUsage() {
		cmdStr := subCmd(cmd).String()
		padding := strings.Repeat(" ", longestCmdLen-len(cmdStr))
		fmt.Fprintf(w, "  %s%s    %s\n", cmdStr, padding, usage)
	}
}
