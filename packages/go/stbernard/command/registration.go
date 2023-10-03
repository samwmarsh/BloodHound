package command

import (
	"github.com/specterops/bloodhound/packages/go/stbernard/command/buildcmd"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/workgen"
)

// subCmd enum represents our subcommands
type subCmd int

const (
	InvalidSubCmd subCmd = iota - 1
	ModSync
	EnvDump
	WorkGen
	Build
)

// String implements Stringer for the Command enum
func (s subCmd) String() string {
	switch s {
	case ModSync:
		return modsync.Name
	case EnvDump:
		return envdump.Name
	case WorkGen:
		return workgen.Name
	case Build:
		return buildcmd.Name
	default:
		return "invalid command"
	}
}

// Commands returns our valid set of Command options
func Commands() []subCmd {
	return []subCmd{ModSync, EnvDump, WorkGen, Build}
}

// Commands usage returns a slice of Command usage statements indexed by their enum
func CommandsUsage() []string {
	var usage = make([]string, len(Commands()))

	usage[ModSync] = modsync.Usage
	usage[EnvDump] = envdump.Usage
	usage[WorkGen] = workgen.Usage
	usage[Build] = buildcmd.Usage

	return usage
}
