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

/// The entrypoint for the generated runner.
pub fn main(call_immediate_fn: CallImmediateFn) {
    let filter =
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info,ftl=debug"));
    fmt::Subscriber::builder().with_env_filter(filter).init();
    info!("Starting runner");

    let cli = Cli::parse();

    debug!(?cli);
}

pub fn build(module_name: &str, src_file: &Path, call_file: &Path, schema_file: &Path) {
    let mut parser = parser::Parser::new();
    let module = parser::ModuleIdent::new(module_name);
    let code = std::fs::read_to_string(src_file).unwrap();
    parser.add_module(&module, &code);
    assert!(
        parser.modules_count() > 0,
        "No modules found in {:?}",
        src_file
    );
    parser.generate_call_immediate_file(call_file);

    let module = parser.generate_module_proto(&module);
    let mut encoded = Vec::new();
    module.encode(&mut encoded).unwrap();
    std::fs::write(schema_file, encoded).unwrap();
}
