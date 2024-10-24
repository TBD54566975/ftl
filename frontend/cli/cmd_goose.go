package main

import (
	"context"
	"fmt"
	"io"
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
		data = "You are working with a system called FTL within an existing project, the docs are at https://raw.githubusercontent.com/michaelneale/ftl-goosehints/refs/heads/main/ftl.txt, please read the docs before doing anything else. You can verify that the project builds via ftl build. Files that end in .ftl.go are auto generated and must not be modified. You can interact with the FTL cluster using the FTL command, run ftl --help for more details. You should not create a new ftl project, module or directory unless explicitly asked. " + data
	} else {
		args = append(args, "--resume-session")
	}
	terminal.UpdateModuleState(ctx, "FTL - Goose", terminal.BuildStateBuilding)
	cmd := exec.Command(ctx, log.Debug, ".", "sq", args...)
	cmd.Stdin = strings.NewReader(data)
	out := &output{}

	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	terminal.UpdateModuleState(ctx, "FTL - Goose", terminal.BuildStateTerminated)
	if err != nil {
		return fmt.Errorf("goose failed: %w", err)
	}
	return nil
}

var _ io.Writer = &output{}

type output struct {
}

func (o output) Write(p []byte) (n int, err error) {
	fmt.Printf("%s", string(p))
	return len(p), nil
}
