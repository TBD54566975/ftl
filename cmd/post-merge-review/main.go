package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

var cli struct {
	Token   string               `env:"GITHUB_TOKEN" required:"true" help:"GitHub token"`
	Webhook kong.FileContentFlag `type:"path" env:"GITHUB_EVENT_PATH" required:"" help:"Path to the webhook payload file."`
}

func main() {
	kctx := kong.Parse(&cli, kong.Description("Create a new issue with all comments from a closed pull request."))

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cli.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	webhook, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	kctx.FatalIfErrorf(err)

	event, err := github.ParseWebHook("issue_comment", webhook)
	kctx.FatalIfErrorf(err)

	commentEvent, ok := event.(*github.IssueCommentEvent)
	if !ok {
		kctx.Fatalf("unexpected event type %T", event)
	}

	pr := commentEvent.GetIssue()
	if pr.GetPullRequestLinks() == nil {
		kctx.Fatalf("comment is not on a pull request")
	}

	if pr.GetState() != "closed" {
		kctx.Fatalf("pull request is not closed")
	}

	prCreator := pr.GetUser().GetLogin()
	repoOwner := commentEvent.GetRepo().GetOwner().GetLogin()
	repo := commentEvent.GetRepo().GetName()

	comments, _, err := client.Issues.ListComments(ctx, repoOwner, repo, pr.GetNumber(), nil)
	kctx.FatalIfErrorf(err)

	checklist := []string{}
	for _, comment := range comments {
		checklist = append(checklist, fmt.Sprintf("- [ ] %s", comment.GetBody()))
	}

	newIssue := &github.IssueRequest{
		Title:     github.String(fmt.Sprintf("Post-merge review from PR #%d", pr.GetNumber())),
		Body:      github.String(strings.Join(checklist, "\n")),
		Labels:    &[]string{"review"},
		Assignees: &[]string{prCreator},
	}

	_, _, err = client.Issues.Create(ctx, repoOwner, repo, newIssue)
	kctx.FatalIfErrorf(err)

	fmt.Println("issue created successfully")
}
