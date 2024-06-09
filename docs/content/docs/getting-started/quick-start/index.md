+++
title = "Quick Start"
description = "One page summary of how to start a new FTL project."
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 20
sort_by = "weight"
template = "docs/page.html"

[extra]
lead = "One page summary of how to start a new FTL project."
toc = true
top = false
+++

## Requirements

### Install the FTL CLI

Install the FTL CLI via [Hermit](https://cashapp.github.io/hermit), [Homebrew](https://brew.sh/), or manually.

#### Hermit (Mac or Linux)

FTL can be installed from the main Hermit package repository by simply:

```
hermit install ftl
```

Alternatively you can add [hermit-ftl](https://github.com/TBD54566975/hermit-ftl) to your sources by adding the following to your Hermit environment's `bin/hermit.hcl` file:

```hcl
sources = ["https://github.com/TBD54566975/hermit-ftl.git", "https://github.com/cashapp/hermit-packages.git"]
```

#### Homebrew (Mac or Linux)

```
brew tap TBD54566975/ftl && brew install ftl
```

#### Manually (Mac or Linux)

Download binaries from the [latest release page](https://github.com/TBD54566975/ftl/releases/latest) and place them in your `$PATH`.

### Install the VSCode extension

The [FTL VSCode extension](https://marketplace.visualstudio.com/items?itemName=FTL.ftl) will run FTL within VSCode, and provide LSP support for FTL, displaying errors within the editor.

## Development

### Create a new module

Once FTL is installed, create a new module:

```
mkdir myproject
cd myproject
ftl init go . alice
```

This will place the code for the new module `alice` in `myproject/alice/alice.go`:

```go
package alice

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
```

Each module is its own Go module.

Any number of modules can be added to your project, adjacent to each other.

### Start the FTL cluster

#### VSCode

If using VSCode, opening the directory will prompt you to start FTL:

[![VSCode](vscode.png)](vscode.png)

#### Manually

Alternatively start the local FTL development cluster from the command-line:

```
ftl dev
```

This will build and deploy all local modules. Modifying the code will cause `ftl
dev` to rebuild and redeploy the module.

### Open the console

FTL has a console that allows navigation of the cluster topology, logs, traces,
and more. Open a browser window at [https://localhost:8892](https://localhost:8892) to view it.

### Create another module

Create another module and call `alice.echo` from it with:

```go
//ftl:verb
import "ftl/alice"

out, err := ftl.Call(ctx, alice.Echo, alice.EchoRequest{})
```

### What next?

Explore the [reference documentation](../../reference/start/).
