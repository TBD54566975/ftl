This Rust runtime implementation needs a fair amount of work to be stable.

Features:
- Verb service and client.
- Can answer verb call requests.
- Can call out to verbs in other runtimes and process responses.
- FTL can build and restart the runtime automatically.
- Generates schema protos for verbs and args/return types recursively.

Limitations:
- Very limited functionality: Verbs only. No generics. Only a few types.
- No error handling (unwraps everywhere).
- No scaffolding generation.
- No external module code generation.
- Only struct types referenced by verb arguments and returns in the same module will correctly work.
- Hardcoded paths to crate dependencies in Cargo.toml
