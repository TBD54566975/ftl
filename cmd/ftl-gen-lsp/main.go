// This program generates hover items for the FTL LSP server.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/tliron/kutil/terminal"
)

type CLI struct {
	Config  string `type:"filepath" default:"internal/lsp/hover.json" help:"Path to the hover configuration file"`
	DocRoot string `type:"dirpath" default:"docs/content/docs" help:"Path to the config referenced markdowns"`
	Output  string `type:"filepath" default:"internal/lsp/hoveritems.go" help:"Path to the generated Go file"`
}

var cli CLI

type hover struct {
	// Match this text for triggering this hover, e.g. "//ftl:typealias"
	Match string `json:"match"`

	// Source file to read from.
	Source string `json:"source"`

	// Select these heading to use for the docs. If omitted, the entire markdown file is used.
	// Headings are included in the output.
	Select []string `json:"select,omitempty"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`Generate hover items for FTL LSP. See lsp/hover.go`),
	)

	hovers, err := parseHoverConfig(cli.Config)
	kctx.FatalIfErrorf(err)

	items, err := scrapeDocs(hovers)
	kctx.FatalIfErrorf(err)

	err = writeGoFile(cli.Output, items)
	kctx.FatalIfErrorf(err)
}

func scrapeDocs(hovers []hover) (map[string]string, error) {
	items := make(map[string]string, len(hovers))
	for _, hover := range hovers {
		path := filepath.Join(cli.DocRoot, hover.Source)
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", path, err)
		}

		doc, err := getMarkdownWithTitle(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		doc.Content = stripNonGoDocs(doc.Content)

		var content string
		if len(hover.Select) > 0 {
			for _, sel := range hover.Select {
				chunk, err := selector(doc.Content, sel)
				if err != nil {
					return nil, fmt.Errorf("failed to select %s from %s: %w", sel, path, err)
				}
				content += chunk
			}
		} else {
			// We need to inject a heading for the hover content because the full content doesn't always have a heading.
			content = fmt.Sprintf("## %s%s", doc.Title, doc.Content)
		}

		items[hover.Match] = content

		// get term width or 80 if not available
		width := 80
		if w, _, err := terminal.GetSize(); err == nil {
			width = w
		}
		line := strings.Repeat("-", width)
		fmt.Print("\x1b[1;34m")
		fmt.Println(line)
		fmt.Println(hover.Match)
		fmt.Println(line)
		fmt.Print("\x1b[0m")
		fmt.Println(content)
	}
	return items, nil
}

func parseHoverConfig(path string) ([]hover, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	var hovers []hover
	err = json.NewDecoder(file).Decode(&hovers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON %s: %w", path, err)
	}

	return hovers, nil
}

type Doc struct {
	Title   string
	Content string
}

// getMarkdownWithTitle reads a Zola markdown file and returns the full markdown content and the title.
func getMarkdownWithTitle(file *os.File) (*Doc, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file.Name(), err)
	}

	// Zola markdown files have a +++ delimiter. An initial one, then metadata, then markdown content.
	parts := bytes.Split(content, []byte("+++"))
	if len(parts) < 3 {
		return nil, fmt.Errorf("file %s does not contain two +++ strings", file.Name())
	}

	// Look for the title in the metadata.
	// title = "PubSub"
	title := ""
	lines := strings.Split(string(parts[1]), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "title = ") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title = "))
			title = strings.Trim(title, "\"")
			break
		}
	}
	if title == "" {
		return nil, fmt.Errorf("file %s does not contain a title", file.Name())
	}

	return &Doc{Title: title, Content: string(parts[2])}, nil
}

var HTMLCommentRegex = regexp.MustCompile(`<!--\w*(.*?)\w*-->`)

// stripNonGoDocs removes non-Go code blocks from the markdown content.
func stripNonGoDocs(content string) string {
	lines := strings.Split(content, "\n")

	var currentLang string
	var collected []string
	for _, line := range lines {
		if strings.Contains(line, "code_selector()") {
			currentLang = ""
			continue
		}
		if line == "{% end %}" {
			currentLang = ""
			continue
		}

		// Grab the language from the HTML comment text
		values := HTMLCommentRegex.FindStringSubmatch(line)
		if len(values) > 1 {
			lang := strings.TrimSpace(values[1])
			// end starts "common" text block
			if lang == "end" {
				currentLang = ""
			} else {
				currentLang = lang
			}
			continue
		}

		// Similar, look at the code block
		if strings.HasPrefix(line, "```") {
			newLang := strings.TrimPrefix(line, "```")
			if newLang != "" {
				currentLang = newLang
			}
		}

		// If we're in a Go block or unspecified, collect the line.
		if currentLang == "go" || currentLang == "" {
			collected = append(collected, line)
		}
	}

	return strings.Join(collected, "\n")
}

func selector(content, selector string) (string, error) {
	// Split the content into lines.
	lines := strings.Split(content, "\n")
	collected := []string{}

	// If the selector starts with ## (the only type of heading we have):
	// Find the line, include it, and all lines until the next heading.
	if !strings.HasPrefix(selector, "##") {
		return "", fmt.Errorf("unsupported selector %s", selector)
	}
	include := false
	for _, line := range lines {
		if include {
			// We have found another heading. Abort!
			if strings.HasPrefix(line, "##") {
				break
			}

			// We also stop at a line break, because we don't want to include footnotes.
			// See the end of docs/content/docs/reference/types.md for an example.
			if line == "---" {
				break
			}

			collected = append(collected, line)
		}

		// Start collecting
		if strings.HasPrefix(line, selector) {
			include = true
			collected = append(collected, line)
		}
	}

	if len(collected) == 0 {
		return "", fmt.Errorf("no content found for selector %s", selector)
	}

	return strings.TrimSpace(strings.Join(collected, "\n")) + "\n", nil
}

func writeGoFile(path string, items map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}
	defer file.Close()

	tmpl, err := template.New("").Parse(`// Code generated by 'just lsp-generate'. DO NOT EDIT.
package lsp

var hoverMap = map[string]string{
{{- range $match, $content := . }}
	{{ printf "%q" $match }}: {{ printf "%q" $content }},
{{- end }}
}
`)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = tmpl.Execute(file, items)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated %s\n", path)
	return nil
}
