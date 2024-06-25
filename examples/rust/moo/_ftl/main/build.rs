use std::env;
use std::path::Path;

fn main() {
    let src = Path::new("../../src/lib.rs");
    let out_dir = env::var("OUT_DIR").unwrap();
    let call_immediate_path = Path::new(&out_dir).join("call_immediate.rs");
    // This does only one module/file
    ftl::runner::build("moo", src, &call_immediate_path);
}