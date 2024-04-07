# FTL VSCode extension

## Getting started

Within the FTL.vscode-workspace, select the `VSCode Extensions` workspace. Select "Run and Debug" on the activity bar, then select `Run Extension`. This will open a new VSCode window with the extension running.

In the extension development host, open an FTL project (with `ftl-project.toml` or `ftl.toml` files).

If you get any errors, you might need to build the extension first (see below).

## Building the extension

We use `just` for our command line tasks. To build the extension, run:
```bash
just build-extension
```

## Packaging the extension

To package the extension, run:
```bash
just package-extension
```

This will create a `.vsix` file in the `extensions/vscode` directory.

## Publishing the extension

To publish the extension, run:
```bash
just publish-extension
```

This will publish the extension to the FTL marketplace. This command will require you to have a Personal Access Token (PAT) with the `Marketplace` scope. You can create a PAT [here](https://dev.azure.com/ftl-org/_usersSettings/tokens).

## Useful links

- [VSCode extension samples](https://github.com/microsoft/vscode-extension-samples)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)
