use std::env;
use std::path::Path;

fn main() {
    let module_dir = Path::new("../..");
    let out_dir = env::var("OUT_DIR").unwrap();

    let src = module_dir.join("src/lib.rs");
    let call_immediate_path = Path::new(&out_dir).join("call_immediate.rs");
    let schema_path = module_dir.join("_ftl").join("schema.pb");

    // This does only one module/file
    ftl::builder::build("moo", &src, &call_immediate_path, &schema_path);
}