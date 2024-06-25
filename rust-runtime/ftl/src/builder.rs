use std::env;
use std::path::Path;

use clap::Parser;
use prost::Message;
use tracing::{debug, info};
use tracing_subscriber::{EnvFilter, fmt};

use crate::parser;
use crate::verb_server::CallImmediateFn;

#[derive(clap::Parser, Debug)]
#[command(version, about, long_about = None)]
struct Cli {
    #[clap(short = 'e', env = "FTL_ENDPOINT", required = true)]
    ftl_endpoint: String,
    #[clap(short = 'c', env = "FTL_CONFIG", required = true)]
    config: Vec<String>,
}

/// The entrypoint for the generated module.
pub fn main(call_immediate_fn: CallImmediateFn) {
    let filter =
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info,ftl=debug"));
    fmt::Subscriber::builder().with_env_filter(filter).init();
    info!("Starting module...");

    let cli = Cli::parse();

    debug!(?cli);
}

pub fn build(module_name: &str) {
    let module_dir = Path::new("..");
    let out_dir = env::var("OUT_DIR").unwrap();

    let src = module_dir.join("src/lib.rs");
    let call_immediate_path = Path::new(&out_dir).join("call_immediate.rs");
    let schema_path = module_dir
        .join("_ftl")
        .join("target")
        .join("debug")
        .join("schema.pb");

    let mut parser = parser::Parser::new();
    let module = parser::ModuleIdent::new(module_name);
    let code = std::fs::read_to_string(&src).unwrap();
    parser.add_module(&module, &code);
    assert!(parser.modules_count() > 0, "No modules found in {:?}", src);
    parser.generate_call_immediate_file(&call_immediate_path);

    let module = parser.generate_module_proto(&module);
    let mut encoded = Vec::new();
    module.encode(&mut encoded).unwrap();
    std::fs::write(&schema_path, &encoded).unwrap();
}
