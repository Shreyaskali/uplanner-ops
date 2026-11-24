package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/shreyaskali/uplanner-ops/internal/update"
)

func runUpdate(root string, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	source := fs.String("source", "", "path to an uplanner-ops source checkout")
	apply := fs.Bool("apply", false, "actually apply the update (default: plan only)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *source == "" {
		return fmt.Errorf("update requires --source PATH; this tool never downloads code by itself")
	}
	if *apply {
		changes, err := update.Apply(root, *source)
		if err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(changes)
		}
		fmt.Println("Applied updates:")
		printChanges(changes)
		return nil
	}
	changes, err := update.Plan(root, *source)
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(changes)
	}
	fmt.Println("Planned updates (nothing written):")
	printChanges(changes)
	fmt.Println("\nRerun with --apply to write. Local edits are preserved as .bak and skipped.")
	return nil
}

func printChanges(changes []update.Change) {
	for _, c := range changes {
		fmt.Printf("  %-10s %-44s %s\n", c.Action, c.Path, c.Detail)
	}
}
