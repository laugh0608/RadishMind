// radishmind-development-probe prepares the offline SDK validation task.
// There is intentionally no serve, model, SDK or command-execution action yet.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"radishmind.local/services/platform/internal/developmentprobe"
)

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: radishmind-development-probe check|prepare|inspect --plan FILE [--source DIR] [--output NEW_DIR] [--candidate DIR]")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	planPath := flags.String("plan", "", "canonical probe plan")
	source := flags.String("source", "", "frozen synthetic task directory")
	output := flags.String("output", "", "new private preparation directory")
	candidate := flags.String("candidate", "", "candidate directory to inspect")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || *planPath == "" {
		return errors.New("probe requires --plan and no positional arguments")
	}
	file, err := os.Open(*planPath)
	if err != nil {
		return fmt.Errorf("read probe plan: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 262144 {
		return errors.New("probe plan must be a bounded regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, 262145))
	if err != nil {
		return fmt.Errorf("read probe plan: %w", err)
	}
	plan, err := developmentprobe.ParsePlan(data)
	if err != nil {
		return err
	}
	var result any
	switch args[0] {
	case "check":
		if *source != "" || *output != "" || *candidate != "" {
			return errors.New("check only accepts --plan")
		}
		result = map[string]any{"status": "plan_valid", "probe_id": plan.ProbeID, "plan_digest": developmentprobe.PlanDigest(plan), "execution": "not_started"}
	case "prepare":
		if *source == "" || *output == "" || *candidate != "" {
			return errors.New("prepare requires --source and --output")
		}
		result, err = developmentprobe.Prepare(plan, *source, *output)
	case "inspect":
		if *candidate == "" || *source != "" || *output != "" {
			return errors.New("inspect requires --candidate")
		}
		result, err = developmentprobe.Inspect(plan, *candidate)
	default:
		return errors.New("unknown probe action; no SDK or execution action is enabled")
	}
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "development-probe:", err)
		os.Exit(1)
	}
}
