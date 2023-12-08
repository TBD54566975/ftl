# Design Doc: JS Interpreter for user codegen

## Motivation

`ftl schema generate` currently relies almost entirely on Go templates for all functionality. For complex types, this means that users have to write recursive inline templates that become quite difficult to reason about. Additionally, most end users aren't going to be Go template experts. The Dart template is an example, with >40 lines of complex recursive templates.

## Goals

Simplify how the template can be extended with complex functions in such a way that most users will find it straightforward.

## Design

Embed the [goja](https://github.com/dop251/goja) JS interpreter and extend the templating support to load a `template.js` file from the template directory. Each top-level function in the JavaScript file will be exposed in the Go template function map.

## Alternatives Considered

Another approach is to use JavaScript for all templating, removing the need for Go templates altogether. The JS VM would have functions for creating files, directories, etc. and ftl would load it and execute it. The JS file would then be responsible for generating everything.

This would be a lot more work for now, vs. the simpler solution above, so we deferred this decision until later.

There's also a middle-ground where, with the proposed
solution, users can create template files with a single
line such as `{{ . | generate }}`
