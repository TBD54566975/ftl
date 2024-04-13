# FTL for Visual Studio Code

[The VS Code FTL extension](https://marketplace.visualstudio.com/items?itemName=FTL.ftl)
provides support for
[FTL](https://github.com/TBD54566975/ftl) within VSCode including LSP integration and useful commands for manageing FTL projects.

## Requirements

- FTL 0.169.0 or newer

## Quick Start

1.  Install [FTL](https://github.com/TBD54566975/ftl) 0.169.0 or newer.

2.  Install this extension.

3.  Open any FTL project with a `ftl-project.toml` or `ftl.toml` file.

4.  The extension will automatically activate and provide support for FTL projects.

> [!IMPORTANT]
> If you have installed FTL with hermit (or other dependency management tools), you may need to specify the path to the FTL binary in the extension settings.

Example:

```json
{
  "ftl.executablePath": "bin/ftl"
}
```

You can also configure additional command line arguments for the FTL binary in the settings.

Example:

```json
{
  "ftl.devCommandFlags": ["--recreate", "--parallelism=4"]
}
```
