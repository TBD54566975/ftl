# FTL README [![CI](https://github.com/TBD54566975/ftl/actions/workflows/ci.yml/badge.svg)](https://github.com/TBD54566975/ftl/actions/workflows/ci.yml)

## Getting started

### Install ftl, for example on macos:
```sh
brew tap TBD54566975/ftl && brew install ftl
```

### Create a sample project (kotlin)
```sh
mkdir myproject
cd myproject
git init .
ftl init kotlin . alice
```

### Serve FTL in a seperate terminal
`ftl serve`

### Deploy and test the module
```sh
ftl deploy ftl-module-alice
ftl call alice.echo '{"name": "Mic"}'
```



### Remember to activate hermit any time you are in the project
```sh
. ./bin/activate-hermit
```

![ftl hacking faster than light](https://github.com/TBD54566975/ftl/assets/14976/37b65b44-021b-4da1-abc2-a5dbcc126c47)




## Project Resources

| Resource                                   | Description                                                                    |
| ------------------------------------------ | ------------------------------------------------------------------------------ |
| [CODEOWNERS](./CODEOWNERS)                 | Outlines the project lead(s)                                                   |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment |
| [CONTRIBUTING.md](./CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues     |
| [GOVERNANCE.md](./GOVERNANCE.md)           | Project governance                                                             |
| [LICENSE](./LICENSE)                       | Apache License, Version 2.0                                                    |
