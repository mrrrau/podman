package generate

import (
	"strconv"
	"strings"

	"github.com/containers/podman/v2/pkg/domain/entities"
	"github.com/pkg/errors"
)

// EnvVariable "PODMAN_SYSTEMD_UNIT" is set in all generated systemd units and
// is set to the unit's (unique) name.
const EnvVariable = "PODMAN_SYSTEMD_UNIT"

// minTimeoutStopSec is the minimal stop timeout for generated systemd units.
// Once exceeded, processes of the services are killed and the cgroup(s) are
// cleaned up.
const minTimeoutStopSec = 60

// RestartPolicies includes all valid restart policies to be used in a unit
// file.
var RestartPolicies = []string{"no", "on-success", "on-failure", "on-abnormal", "on-watchdog", "on-abort", "always"}

// validateRestartPolicy checks that the user-provided policy is valid.
func validateRestartPolicy(restart string) error {
	for _, i := range RestartPolicies {
		if i == restart {
			return nil
		}
	}
	return errors.Errorf("%s is not a valid restart policy", restart)
}

func generateHeaderTemplate(options entities.GenerateSystemdOptions) string {
	const serviceHeader = `# {{{{.ServiceName}}}}.service
`
	const infoHeader = `# autogenerated by Podman {{{{.PodmanVersion}}}}
{{{{- if .TimeStamp}}}}
# {{{{.TimeStamp}}}}
{{{{- end}}}}
`
	const unitHeader = `[Unit]
Description=Podman {{{{.ServiceName}}}}.service
Documentation=man:podman-generate-systemd(1)
Wants=network.target
After=network-online.target
`
	template := serviceHeader + infoHeader
	if options.NoHeader {
		template = serviceHeader
	}
	template = template + "\n" + unitHeader

	return template
}

// filterPodFlags removes --pod and --pod-id-file from the specified command.
func filterPodFlags(command []string) []string {
	processed := []string{}
	for i := 0; i < len(command); i++ {
		s := command[i]
		if s == "--pod" || s == "--pod-id-file" {
			i++
			continue
		}
		if strings.HasPrefix(s, "--pod=") || strings.HasPrefix(s, "--pod-id-file=") {
			continue
		}
		processed = append(processed, s)
	}
	return processed
}

// quoteArguments makes sure that all arguments with at least one whitespace
// are quoted to make sure those are interpreted as one argument instead of
// multiple ones.
func quoteArguments(command []string) []string {
	for i := range command {
		if strings.ContainsAny(command[i], " \t") {
			command[i] = strconv.Quote(command[i])
		}
	}
	return command
}

func removeDetachArg(args []string, argCount int) []string {
	// "--detach=false" could also be in the container entrypoint
	// split them off so we do not remove it there
	realArgs := args[len(args)-argCount:]
	flagArgs := removeArg("-d=false", args[:len(args)-argCount])
	flagArgs = removeArg("--detach=false", flagArgs)
	return append(flagArgs, realArgs...)
}

func removeReplaceArg(args []string, argCount int) []string {
	// "--replace=false" could also be in the container entrypoint
	// split them off so we do not remove it there
	realArgs := args[len(args)-argCount:]
	flagArgs := removeArg("--replace=false", args[:len(args)-argCount])
	return append(flagArgs, realArgs...)
}

func removeArg(arg string, args []string) []string {
	newArgs := []string{}
	for _, a := range args {
		if a != arg {
			newArgs = append(newArgs, a)
		}
	}
	return newArgs
}
