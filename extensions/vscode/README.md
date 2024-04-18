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

4.  The extension will prompt to start the FTL development server.

## Settings

Configure the FTL extension by setting the following options in your Visual Studio Code settings.json:

- `ftl.executablePath`: Specifies the path to the FTL executable. The default is "ftl", which uses the system's PATH to find the executable.

  > [!IMPORTANT]
  > If you have installed FTL with hermit (or other dependency management tools), you may need to specify the path to the FTL binary in the extension settings.

  ```json
  {
    "ftl.executablePath": "bin/ftl"`
  }
  ```

- `ftl.devCommandFlags`: Defines flags to pass to the FTL executable when starting the development environment. The default is ["--recreate"].

  ```json
  {
    "ftl.devCommandFlags": ["--recreate", "--parallelism=4"]
  }
  ```

- `ftl.startClientOption`: Controls if and when to automatically start the FTL development server. Available options are "always" and "never". If not set, the extension will prompt to start the server when opening a FTL project.
  ```json
  {
    "ftl.startClientOption": "always"
  }
  ```
