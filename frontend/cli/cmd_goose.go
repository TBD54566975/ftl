package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/terminal"
)

type gooseCmd struct {
	Prompt []string `arg:"" required:"" help:"The goose prompt"`
}

var first = true

func (c *gooseCmd) Run(ctx context.Context) error {
	data := strings.Join(c.Prompt, " ")
	args := []string{"goose", "run"}
	if first {
		first = false
		data = "You are working with a system called FTL within an existing project, the docs are at https://github.com/TBD54566975/ftl/tree/main/docs/content/docs/reference. You can verify that the project builds via ftl build. Files that end in .ftl.go are auto generated and must not be modified." + data
	} else {
		args = append(args, "--resume-session")
	}
	terminal.UpdateModuleState(ctx, "FTL - Goose", terminal.BuildStateBuilding)
	cmd := exec.Command(ctx, log.Debug, ".", "sq", args...)
	cmd.Stdin = strings.NewReader(data)
	cmd.Stdout = nil
	cmd.Stderr = nil
	capture, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", capture)
	terminal.UpdateModuleState(ctx, "FTL - Goose", terminal.BuildStateTerminated)
	if err != nil {
		return fmt.Errorf("goose failed: %w", err)
	}
	return nil
}
