# FTL modules

Each subdirectory represents an FTL module. Remote modules will be
code-generated into their own directories, one module per directory. Note that
this is a temporary solution.

For example given an `echo` module written in Go that calls a `time` module
written in Kotlin, the filesystem might look like this once the FTL tooling is
started:

```
README.md
go.mod
echo/ftl.toml
echo/echo.go
time/generated_ftl_module.go
```
