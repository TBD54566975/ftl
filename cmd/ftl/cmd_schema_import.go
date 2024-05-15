package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/alecthomas/types/optional"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/log"
)

type schemaImportCmd struct {
	OllamaPort int             `help:"Port to use for ollama." default:"11434"`
	Go         importGoCmd     `cmd:"" help:"Import types for an FTL Go module."`
	Kotlin     importKotlinCmd `cmd:"" help:"Import types for an FTL Kotlin module."`
}

type importGoCmd struct {
	Dir string `arg:"" required:"" help:"Directory to import from."`
}

type importKotlinCmd struct {
	Dir string `arg:"" required:"" help:"Directory to import from."`
}

const ollamaContainerName = "ftl-ollama-1"
const ollamaVolume = "ollama:/root/.ollama"
const ollamaModel = "llama2"

func (i *importKotlinCmd) getPrompt() (string, error) {
	input, err := createInputString(i.Dir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("I will provide you with an input string containing the contents of a file which must be translated into Kotlin. "+
		"The resulting Kotlin file should translate every object from the input into its own Kotlin data class, written as `data class...`. "+
		"The data class parameters should each be declared as `val`."+
		"Everything in the provided input that is not a type or object declaration must be ignored."+
		"No fields should be modified or added—the result must be an exact translation of the provided input, retaining case."+
		"The result should reside in a package named `ftl`."+
		"Your output should only include the resultant Kotlin data classes with no additional functions."+
		"Please provide output that is only the resultant Kotlin file, with no additional text or explanation. The initial string to translate is this: %s",
		input,
	), nil
}

func (i *importGoCmd) getPrompt() (string, error) {
	input, err := createInputString(i.Dir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("I will provide you with an input string containing the contents of a file which must be translated into Go. "+
		"The resulting Go file should represent every object as its own Go struct, written as `type...`. "+
		"Everything in the provided input that is not a type or object declaration must be ignored."+
		"No fields should be modified or added—the result must be an exact translation of the provided input, retaining case."+
		"The result should reside in a package named `ftl`."+
		"Your output should only include the resultant Go types with no additional functions."+
		"Please provide output that is only the resultant Go file, with no additional text or explanation. The initial string to translate is this: %s",
		input,
	), nil
}

func (i *importGoCmd) Run(ctx context.Context, parent *schemaImportCmd) error {
	err := parent.setup(ctx)
	if err != nil {
		return err
	}

	prompt, err := i.getPrompt()
	if err != nil {
		return err
	}

	err = query(ctx, prompt)
	if err != nil {
		return err
	}

	return nil
}

func (i *importKotlinCmd) Run(ctx context.Context, parent *schemaImportCmd) error {
	err := parent.setup(ctx)
	if err != nil {
		return err
	}

	prompt, err := i.getPrompt()
	if err != nil {
		return err
	}

	err = query(ctx, prompt)
	if err != nil {
		return err
	}

	return nil
}

func query(ctx context.Context, prompt string) error {
	logger := log.FromContext(ctx)

	logger.Debugf("The import schema command relies on the %s AI chat model to translate schemas from other "+
		"languages into FTL compliant objects in the specified language. Output may vary and results should be inspected "+
		"for correctness. It is suggested that if the results are not satisfactory, you try again.\n\n", ollamaModel)

	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		logger.Errorf(err, "failed to call ollama")
	}

	completion, err := llm.Call(ctx, fmt.Sprintf("Human: %s \nAssistant:", prompt),
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		logger.Errorf(err, "failed to call ollama")
	}

	_ = completion
	return nil
}

func (s *schemaImportCmd) setup(ctx context.Context) error {
	logger := log.FromContext(ctx)

	exists, err := container.DoesExist(ctx, ollamaContainerName)
	if err != nil {
		return err
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for ollama", ollamaContainerName)

		// check if port s.OllamaPort is already in use
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.OllamaPort))
		if err != nil {
			return fmt.Errorf("port %d is already in use: %w", s.OllamaPort, err)
		}
		_ = l.Close()

		err = container.Run(ctx, "ollama/ollama", ollamaContainerName, s.OllamaPort, 11434, optional.Some(ollamaVolume))
		if err != nil {
			return err
		}
	} else {
		// Start the existing container
		err = container.Start(ctx, ollamaContainerName)
		if err != nil {
			return err
		}
	}

	// Initialize Ollama
	err = container.Exec(ctx, ollamaContainerName, "ollama", "run", ollamaModel)
	if err != nil {
		return err
	}

	return nil
}

func createInputString(dir string) (string, error) {
	var result string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			result += string(fileContent) + "\n"
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return result, nil
}
