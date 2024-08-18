package main

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/kong"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CLI struct {
	MaxParallel int      `short:"j" help:"Maximum number of tests to run in parallel." default:"2"`
	Tests       []string `arg:"" help:"Tests to run. If not specified, all tests will be run." optional:""`
	Nocache     bool     `help:"Do not use cache when building docker image."`
}

type job struct {
	testPath string
	testName string
}

type result struct {
	job    job
	err    error
	output string
	taken  time.Duration
}

// TestDiskSchemaRetrieverWithBuildArtefact TestCron TestFSM TestPubSub TestDatabase TestLeaderSync

const dockerComposeTemplate = `
services:
  db:
    image: postgres:15.8
    command: postgres
    user: postgres
    restart: always
    environment:
      POSTGRES_PASSWORD: secret
    healthcheck:
      test: ["CMD-SHELL", "pg_isready"]
      interval: 1s
      timeout: 60s
      retries: 60
      start_period: 80s
  localstack:
    image: localstack/localstack
    environment:
      SERVICES: secretsmanager
      DEBUG: 1
  ftl:
    image: ftl-integration-test
    depends_on:
      - db
      - localstack
    platform: linux/amd64
    environment:
      FTL_DATABASE_IMAGE: "none"
    volumes:
      - $integrationCache/go:/root/go
      - $integrationCache/.m2:/root/.m2
      - $integrationCache/java-runtime-test-framework:/ftl/java-runtime/ftl-runtime/test-framework/target
      - $integrationCache/java-runtime-target:/ftl/java-runtime/ftl-runtime/target
      - $workerCache/.cache:/root/.cache
      - $workerCache/.npm:/root/.npm
      - $workerCache/frontend-node-modules:/ftl/frontend/node_modules
    #command: bash -c "/root/run.sh $testName"
    command: bash -c "sleep infinity"
`

const dockerfileFTL = `
FROM ubuntu:24.04
RUN apt-get update
RUN apt-get install -y curl git zip postgresql-client iputils-ping vim
ENV PATH $PATH:/root/bin:/ftl/bin
RUN curl -fsSL https://github.com/cashapp/hermit/releases/download/stable/install.sh | /bin/bash
COPY ./bin /ftl/bin
RUN go version
RUN mvn -f kotlin-runtime/ftl-runtime -B --version
RUN node --version
RUN echo #!/bin/bash > /root/run.sh
RUN echo "socat TCP-LISTEN:15432,fork TCP:db:5432 &" >> /root/run.sh
RUN echo "socat TCP-LISTEN:4566,fork TCP:localstack:4566 &" >> /root/run.sh
RUN echo "if [ -z \"\$1\" ]; then" >> /root/run.sh
RUN echo "  echo no test specified" >> /root/run.sh
RUN echo "  exit 1" >> /root/run.sh
RUN echo fi >> /root/run.sh
RUN echo "just integration-tests \"\$1\"" >> /root/run.sh
RUN apt-get install -y socat
RUN chmod +x /root/run.sh
COPY . /ftl
WORKDIR /ftl
`

// This is a way of running the integration tests in parallel inside docker-compose sets, so that the DB, localstack are all separate.
// The volumes will mount from the host machine to the container, so we don't need to rebuild constantly.
func main() {
	var cli CLI
	kctx := kong.Parse(&cli)
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	logger := log.FromContext(ctx)

	gitRoot, ok := internal.GitRoot("").Get()
	if !ok {
		kctx.FatalIfErrorf(fmt.Errorf("failed to find Git root"))
	}

	// make a temporary directory to store the docker files
	tempDir, err := os.MkdirTemp("", "ftl-integration-test-runner")
	kctx.FatalIfErrorf(err)
	defer os.RemoveAll(tempDir)

	// write the dockerfile
	err = os.WriteFile(tempDir+"/Dockerfile", []byte(dockerfileFTL), 0644)
	kctx.FatalIfErrorf(err)

	dockerStart := time.Now()
	// build the docker image
	logger.Infof("Building docker image")
	args := []string{"build"}
	if cli.Nocache {
		args = append(args, "--no-cache")
	}
	args = append(args,
		"--platform=linux/amd64",
		"-f", tempDir+"/Dockerfile",
		"-t", "ftl-integration-test",
		".")

	err = exec.Command(ctx, log.Info, gitRoot, "docker", args...).RunBuffered(ctx)
	kctx.FatalIfErrorf(err, "failed to build docker image... maybe try with --no-cache ?")

	logger.Infof("Docker image built in %s", time.Since(dockerStart))

	start := time.Now()

	tests, err := getTests(ctx, tempDir, cli.Tests)
	kctx.FatalIfErrorf(err)

	jobs := make(chan job, len(tests))
	results := make(chan result, len(tests))

	for i := 0; i < cli.MaxParallel; i++ {
		go worker(ctx, tempDir, jobs, i, results)
	}

	for _, test := range tests {
		logger.Debugf("Adding test %s to queue", test)
		jobs <- test
	}
	close(jobs)

	// immediately stop if there is a result
	for i := 0; i < len(tests); i++ {
		result := <-results
		logger.Infof("Test %s took %s err: %v", result.job.testName, result.taken, result.err)
		if result.err != nil {
			cancel()
			logger.Infof(result.output)
			kctx.FatalIfErrorf(result.err)
		}
	}

	logger.Infof("All tests passed in %s", time.Since(start))
}

// getTests returns a list of tests to run. If no tests are specified, all tests will be returned.
// Recursive search for all files that have the _test.go suffix and contain the //go:build integration tag.
func getTests(ctx context.Context, tempDir string, requested []string) ([]job, error) {
	logger := log.FromContext(ctx)

	var tests []job
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if !strings.Contains(string(data), "//go:build integration") {
			return nil
		}

		re := regexp.MustCompile(`func (Test\w*)`)
		matches := re.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			j := job{testPath: path, testName: match[1]}
			tests = append(tests, j)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk: %w", err)
	}
	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests found at all")
	}

	// filter by requested
	if len(requested) > 0 {
		logger.Infof("Filtering tests by requested: %v", requested)
		logger.Infof("len(requested) = %d", len(requested))
		var filtered []job
		for _, test := range tests {
			for _, req := range requested {
				if test.testName == req {
					filtered = append(filtered, test)
				}
			}
		}
		if len(filtered) == 0 {
			logger.Warnf("No tests found for %v", requested)
			logger.Warnf("Available tests: %v", slices.Map(tests, func(j job) string { return j.testName }))
			return nil, fmt.Errorf("no tests found for %v", requested)
		}
		tests = filtered
	}

	return tests, nil
}

func worker(ctx context.Context, tempDir string, jobs <-chan job, id int, results chan<- result) {
	logger := log.FromContext(ctx)
	jobName := fmt.Sprintf("worker-%d", id)

	gitRoot, ok := internal.GitRoot("").Get()
	if !ok {
		panic("failed to find Git root")
	}

	composePath, err := generateCompose(tempDir, jobName, gitRoot, id)
	if err != nil {
		panic("failed to write docker-compose file")
	}

	composeProjectName := strcase.ToLowerKebab(jobName)
	args := []string{"-f", composePath,
		"--project-name", composeProjectName,
		"--project-directory", gitRoot,
	}

	start := time.Now()
	ftlArgs := append(args, "up", "-d", "ftl")
	out, err := exec.Capture(ctx, gitRoot, "docker-compose", ftlArgs...)
	if err != nil {
		panic(fmt.Sprintf("failed to run docker-compose: %v %s", err, out))
	}

	logger.Debugf("Worker %s started in %s", jobName, time.Since(start))

	for job := range jobs {
		logger.Infof("Running test %s", job.testName)
		start = time.Now()

		// use run
		args := append(args, "exec", "ftl", "bash", "-c", fmt.Sprintf("/root/run.sh %s", job.testName))
		//out, err := exec.Capture(ctx, gitRoot, "docker-compose", args...)
		err := exec.Command(ctx, log.Info, gitRoot, "docker-compose", args...).Run()
		if err != nil {
			results <- result{job: job, err: err, output: string(out)}
			continue
		}

		results <- result{job: job, taken: time.Since(start), output: string(out)}
	}

	logger.Debugf("Shutting down %s", composeProjectName)
	err = exec.Command(ctx, log.Trace, gitRoot, "docker-compose", append(args, "down")...).Run()
	if err != nil {
		panic(fmt.Sprintf("failed to shut down docker-compose: %v", err))
	}
}

func generateCompose(tempDir, suffix, gitRoot string, workerId int) (string, error) {
	composePath := filepath.Join(tempDir, fmt.Sprintf("docker-compose-%s.yaml", suffix))
	content := strings.ReplaceAll(dockerComposeTemplate, "$testName", suffix)
	content = strings.ReplaceAll(content, "$projectRoot", gitRoot)
	content = strings.ReplaceAll(content, "$tempDir", tempDir)
	content = strings.ReplaceAll(content, "$userHome", os.Getenv("HOME"))
	integrationCache := fmt.Sprintf("%s/.cache/ftl/integration-tests/shared", os.Getenv("HOME"))
	content = strings.ReplaceAll(content, "$integrationCache", integrationCache)
	workerCache := fmt.Sprintf("%s/.cache/ftl/integration-tests/worker-%d", os.Getenv("HOME"), workerId)
	content = strings.ReplaceAll(content, "$workerCache", workerCache)
	err := os.WriteFile(composePath, []byte(content), 0644)
	return composePath, err
}
