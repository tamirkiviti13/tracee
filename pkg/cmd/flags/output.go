package flags

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquasecurity/tracee/pkg/cmd/printer"
	tracee "github.com/aquasecurity/tracee/pkg/ebpf"
)

func outputHelp() string {
	return `Control how and where output is printed.
Possible options:
[format:]table                                     output events in table format
[format:]table-verbose                             output events in table format with extra fields per event
[format:]json                                      output events in json format
[format:]gob                                       output events in gob format
[format:]gotemplate=/path/to/template              output events formatted using a given gotemplate file
out-file:/path/to/file                             write the output to a specified file. create/trim the file if exists (default: stdout)
log-file:/path/to/file                             write the logs to a specified file. create/trim the file if exists (default: stderr)
none                                               ignore stream of events output, usually used with --capture
option:{stack-addresses,exec-env,relative-time,exec-hash,parse-arguments,sort-events}
                                                   augment output according to given options (default: none)
  stack-addresses                                  include stack memory addresses for each event
  exec-env                                         when tracing execve/execveat, show the environment variables that were used for execution
  relative-time                                    use relative timestamp instead of wall timestamp for events
  exec-hash                                        when tracing sched_process_exec, show the file hash(sha256) and ctime
  parse-arguments                                  do not show raw machine-readable values for event arguments, instead parse into human readable strings
  parse-arguments-fds                              enable parse-arguments and enrich fd with its file path translation. This can cause pipeline slowdowns.
  sort-events                                      enable sorting events before passing to them output. This will decrease the overall program efficiency.
  cache-events                                     enable caching events to release perf-buffer pressure. This will decrease amount of event loss until cache is full.
Examples:
  --output json                                            | output as json
  --output gotemplate=/path/to/my.tmpl                     | output as the provided go template
  --output out-file:/my/out --output log-file:/my/log      | output to /my/out and logs to /my/log
  --output none                                            | ignore events output
Use this flag multiple times to choose multiple output options
`
}

type OutputConfig struct {
	tracee.OutputConfig
	LogPath string
	LogFile *os.File
}

func PrepareOutput(outputSlice []string) (OutputConfig, printer.Config, error) {
	outcfg := OutputConfig{}
	printcfg := printer.Config{}
	printerKind := "table"
	outPath := ""
	for _, o := range outputSlice {
		outputParts := strings.SplitN(o, ":", 2)
		numParts := len(outputParts)
		if numParts == 1 && outputParts[0] != "none" {
			outputParts = append(outputParts, outputParts[0])
			outputParts[0] = "format"
		}

		switch outputParts[0] {
		case "none":
			printerKind = "ignore"
		case "format":
			printerKind = outputParts[1]
			if printerKind != "table" &&
				printerKind != "table-verbose" &&
				printerKind != "json" &&
				printerKind != "gob" &&
				!strings.HasPrefix(printerKind, "gotemplate=") {
				return outcfg, printcfg, fmt.Errorf("unrecognized output format: %s. Valid format values: 'table', 'table-verbose', 'json', 'gob' or 'gotemplate='. Use '--output help' for more info", printerKind)
			}
		case "out-file":
			outPath = outputParts[1]
		case "log-file":
			outcfg.LogPath = outputParts[1]
		case "option":
			switch outputParts[1] {
			case "stack-addresses":
				outcfg.StackAddresses = true
			case "exec-env":
				outcfg.ExecEnv = true
			case "relative-time":
				outcfg.RelativeTime = true
				printcfg.RelativeTS = true
			case "exec-hash":
				outcfg.ExecHash = true
			case "parse-arguments":
				outcfg.ParseArguments = true
			case "parse-arguments-fds":
				outcfg.ParseArgumentsFDs = true
				outcfg.ParseArguments = true // no point in parsing file descriptor args only
			case "sort-events":
				outcfg.EventsSorting = true
			default:
				return outcfg, printcfg, fmt.Errorf("invalid output option: %s, use '--output help' for more info", outputParts[1])
			}
		default:
			return outcfg, printcfg, fmt.Errorf("invalid output value: %s, use '--output help' for more info", outputParts[1])
		}
	}

	if printerKind == "table" {
		outcfg.ParseArguments = true
	}

	printcfg.Kind = printerKind

	if outPath == "" {
		printcfg.OutFile = os.Stdout
	} else {
		printcfg.OutPath = outPath
		fileInfo, err := os.Stat(outPath)
		if err == nil && fileInfo.IsDir() {
			return outcfg, printcfg, fmt.Errorf("cannot use a path of existing directory %s", outPath)
		}
		dir := filepath.Dir(outPath)
		os.MkdirAll(dir, 0755)
		printcfg.OutFile, err = os.Create(outPath)
		if err != nil {
			return outcfg, printcfg, fmt.Errorf("failed to create output path: %v", err)
		}
	}

	if outcfg.LogPath == "" {
		outcfg.LogFile = os.Stderr
	} else {
		errPath := outcfg.LogPath
		fileInfo, err := os.Stat(errPath)
		if err == nil && fileInfo.IsDir() {
			return outcfg, printcfg, fmt.Errorf("cannot use a path of existing directory %s", errPath)
		}
		dir := filepath.Dir(errPath)
		os.MkdirAll(dir, 0755)
		outcfg.LogFile, err = os.Create(errPath)
		if err != nil {
			return outcfg, printcfg, fmt.Errorf("failed to create output path: %v", err)
		}
	}

	return outcfg, printcfg, nil
}
