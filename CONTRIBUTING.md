\*\*\*\* Contribution Guide

There are many ways to be an open source contributor, and we're here to help you on your way! You may:

- Raise an issue or feature request in our [issue tracker](https://github.com/TBD54566975/ftl/issues)
- Help another contributor with one of their questions, or a code review
- Suggest improvements to our Getting Started documentation by supplying a Pull Request
- Evangelize our work together in conferences, podcasts, and social media spaces.

This guide is for you.

## Development Prerequisites

We recommend that you use OrbStack instead of Docker desktop when developing on this project:

```
brew install orbstack
```

or [OrbStack Website](https://orbstack.dev/)

The tools used by this project are managed by
[Hermit](https://cashapp.github.io/hermit/), a self-bootstrapping package
installer. To activate the Hermit environment, cd into the source directory and
type:

```
$ . ./bin/activate-hermit
```

> **Tip:** Install Hermit's [shell hooks](https://cashapp.github.io/hermit/usage/shell/) to automatically activate Hermit environments when you `cd` into them.

Once Hermit has been activated you need to install all required dependencies. To do this run:

```
$ hermit install
```

If using an IDEA IDE we highly recommend installing
the Hermit [IDEA plugin](https://plugins.jetbrains.com/plugin/16882-hermit).
For most other editors, starting your editor from within a terminal in an
activated Hermit environment is the best way to ensure the editor's environment
variables are correctly set. eg. for VSCode:

```
$ . ./bin/activate-hermit
$ code .
```

Once the Hermit environment is activated you can type the following to start a
hot-reloading ftl agent:

```
$ ftl dev --recreate ./examples/go
```

## Development processes

### Code reviews

Our goal is to maintain velocity while keeping code quality high. To that end, the process is an exercise in trust on the part of the reviewer, and responsibility on the part of the PR author.

In practice, the **reviewer** will review _and_ approve the PR at the same time, trusting that the author will apply the feedback before merging.

On the **author's** side, they are responsible for reading and understanding all feedback, and applying that feedback where they think it is appropriate. If either side doesn't understand something and it's important, comment accordingly, or do a quick pairing to resolve it. The author should feel free to re-request a review.

Additional points of note:

- We discourage bike-shedding. Code and documentation are easy to change, we can always adjust it later.
- Keep your PRs digestible, large PRs are very difficult to comprehend and review.
- Changing code is cheap, we can fix it later. The only caveat here is data storage.
- Reviewing code is everybody's responsibility.

### Design process

All of our design documents live in [HackMD](https://hackmd.io/@ftl). Create new design documents from the [FTL Design Doc Template](https://hackmd.io/team/ftl?nav=overview&template=e98dd2c1-636d-402d-95a9-4dba73ca333a).

Many changes are relatively straightforward and don't require a design, but more complex changes should be covered by a design document. Once a DRI (Directly Responsible Individual, a Cash term) is selected, our process for creating and reviewing designs is the following. The DRI:

1. Does some initial thinking and comes up with a rough approach.
2. Discuss the design with the team, and particularly the DRI for the subsystem that the change will affect.
3. If the design seems broadly sound and useful, continue on to write the design doc.
4. A day or two before the weekly sync, send out the design doc for review.
5. Address any comments by changing the design or discussion.
6. Add the design review to the weekly sync meeting. The sync meeting is a good opportunity to discuss the design with the team and address any last feedback.
7. Once approved by the DRI for that subsystem and the DRI for FTL (if necessary), implement!

**Reviewers should avoid bikeshedding.**

Of course, feel free to bounce ideas off anyone on the team at any time.

Our design docs are [stored in HackMD](https://hackmd.io/team/ftl). There's a template for design docs that should be used. Add a label to your design document, representing its state.

## Coding best practices

### Optional

We prefer to use [types.Optional\[T\]](https://pkg.go.dev/github.com/alecthomas/types/optional) as opposed to `nil` pointers for expressing that a value may be missing. This is because pointers are semantically ambiguous and error prone. They can mean, variously: "this value may or may not be present", "this value just happens to be a pointer", or "this value is a pointer because it's mutable". Using a `types.Optional[T]`, even for pointers, expresses the intent much more clearly.

### Sum types

We're using a tool called go-check-sumtype, which is basically a hacky way to
check that switch statements on certain types are exhaustive. An example of that
is schema.Type. The way it works is that if you have a type switch like this:

```go
switch t := t.(type) {
  case *schema.Int:
}
```

It will detect that the switch isn't exhaustive. However, if you have this:

```go
switch t := t.(type) {
  case *schema.Int:
  default:
    return fmt.Errorf("unsupported type")
}
```

Then it will assume you're intentionally handling the default case, and won't
detect it. Instead, if you panic in the default case rather than returning an error,
`go-check-sumtype` will be able to detect the missing case. A panic is usually
what we want anyway, because it isn't a recoverable error.

TL;DR Don't do the above, do this:

```go
switch t := t.(type) {
  case *schema.Int:
    // Handle Int

  // For all cases you don't want to handle, enumerate them with an
  // empty case.
  case *schema.String, *schema.Bool /* etc. */:
    // Do nothing
}
```

Then when a new case is added to the sum type, `go-check-sumtype` will detect the missing case statically.

### Database and SQL changes

If you make any changes to the `sqlc` inputs, i.e. all the `sql/queries.sql` files, the contents of `backend/controller/sql/schema`, or `sqlc.yaml`, then you will need to update the Go code that `sqlc` generates from those inputs:

```bash
just build-sqlc
```

We use [dbmate](https://github.com/amacneil/dbmate) to manage migrations. To create a migration file, run `dbmate new` with the name of your migration. Example:

```
dbmate new create_users_table
```

This will automatically create a migration file in `backend/controller/sql/schema/`. You can refer to any of the existing files in there as examples while writing your own migration.

[This section](https://github.com/amacneil/dbmate?tab=readme-ov-file#creating-migrations) of the dbmate docs explains how to create a migration if you'd like to learn more.

## VSCode extension

The preferred way to develop the FTL VSCode extension is to open a VSCode instance in the `frontend/vscode` directory. This will load the extension in a new VSCode window. From there, the `launch.json` and `tasks.json` files are configured to run the extension in a new window.

### Building the extension

```bash
just build-extension
```

### Packaging the extension

To package the extension, run:

```bash
just package-extension
```

This will create a `.vsix` file in the `frontend/vscode` directory.

### Publishing the extension

To publish the extension, run:

```bash
just publish-extension
```

This will publish the extension to the FTL marketplace. This command will require you to have a Personal Access Token (PAT) with the `Marketplace` scope. You can create a PAT [here](https://dev.azure.com/ftl-org/_usersSettings/tokens).

## Debugging with Delve

### Building with Full Debug Information

To build a binary with full debug information for Delve, use the following command:

```sh
FTL_DEBUG=true just build ftl
```

### Debugging a Running Process

For an in-line replacement of `ftl dev <args>`, use the command:

```sh
just debug <args>
```

This command compiles a binary with debug information, runs `ftl dev <args>` using this binary, and provides an endpoint to attach a remote debugger at **127.0.0.1:2345**.
You do not need to run `FTL_DEBUG=true just build ftl` separately when using this command.

### Attaching a Debugger

By running `just debug <args>` and then attaching a remote debugger, you can debug the FTL infrastructure while running your project.

#### IntelliJ

Run `Debug FTL` from the `Run/Debug Configurations` dropdown while in the FTL project.

#### VSCode

Run `Debug FTL` from the `Run and Debug` dropdown while in the FTL project.

## Running with Grafana for Metrics and Tracing

Start the Grafana stack with:

```sh
just grafana
```

This will start Grafana (can be stopped via `just grafana-stop`). You can access Grafana at [http://localhost:3000](http://localhost:3000) with the default credentials `admin:admin`.

Once Grafana is running, you can start FTL with otel enabled:

```sh
just otel-dev
```

## Testing local changes

To use your locally built FTL in a separate project, you can start live rebuild with
```sh
just live-rebuild
```

Then, in a separate terminal, you can use the locally built FTL to test your changes against a separate FTL project by running the locally built FTL from the root of this project:
```sh
${FTL_HOME}/build/release/ftl dev
```
where `FTL_HOME` is the root of this repository.

## Useful links

- [VSCode extension samples](https://github.com/microsoft/vscode-extension-samples)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)

## Communications

### Issues

Anyone from the community is welcome (and encouraged!) to raise issues via
[GitHub Issues](https://github.com/TBD54566975/ftl/issues).

We have an [automated aggregation issue](https://github.com/TBD54566975/ftl/issues/728) that lists all the PRs and issues people are working on.

### Continuous Integration

Build and Test cycles are run on every commit to every branch on [GitHub Actions](https://github.com/TBD54566975/ftl/actions).

## Contribution

We review contributions to the codebase via GitHub's Pull Request mechanism. We have
the following guidelines to ease your experience and help our leads respond quickly
to your valuable work:

- Start by proposing a change either in Issues (most appropriate for small
  change requests or bug fixes) or in Discussions (most appropriate for design
  and architecture considerations, proposing a new feature, or where you'd
  like insight and feedback)
- Cultivate consensus around your ideas; the project leads will help you
  pre-flight how beneficial the proposal might be to the project. Developing early
  buy-in will help others understand what you're looking to do, and give you a
  greater chance of your contributions making it into the codebase! No one wants to
  see work done in an area that's unlikely to be incorporated into the codebase.
- Fork the repo into your own namespace/remote
- Work in a dedicated feature branch. Atlassian wrote a great
  [description of this workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/feature-branch-workflow)
- When you're ready to offer your work to the project, first:
- Squash your commits into a single one (or an appropriate small number of commits), and
  rebase atop the upstream `main` branch. This will limit the potential for merge
  conflicts during review, and helps keep the audit trail clean. A good writeup for
  how this is done is
  [here](https://medium.com/@slamflipstrom/a-beginners-guide-to-squashing-commits-with-git-rebase-8185cf6e62ec), and if you're
  having trouble - feel free to ask a member or the community for help or leave the commits as-is, and flag that you'd like
  rebasing assistance in your PR! We're here to support you.
- Open a PR in the project to bring in the code from your feature branch. We use
  [Conventional Commits](https://www.conventionalcommits.org/), so prefix your PR title
  with `feat:`, `fix:`, or `chore:` appropriately. `feat` will result in a minor version
  bump, and `fix` will result in a patch version bump.
- The maintainers noted in the `CODEOWNERS` file will review your PR and optionally
  open a discussion about its contents before moving forward.
- Remain responsive to follow-up questions, be open to making requested changes, and...
  You're a contributor!
- And remember to respect everyone in our global development community. Guidelines
  are established in our `CODE_OF_CONDUCT.md`.
