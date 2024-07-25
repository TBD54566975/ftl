use std::fs::File;
use std::io::Write;
use std::path::PathBuf;

use clap::{Parser, Subcommand};
use prost::Message;
use tracing::{error, info};

use ftl::parser;

#[derive(Parser)]
#[command(version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    BuildSchema { input: PathBuf, output: PathBuf },
    DumpModuleSchema { file: PathBuf },
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    tracing_subscriber::fmt::init();

    match cli.command {
        Some(Commands::BuildSchema { input, output }) => {
            info!("Building schema from {:?}", input);
            let name = input.file_stem().unwrap().to_str().unwrap();
            let content = std::fs::read_to_string(&input).unwrap();

            let mut parser = parser::Parser::new();
            let module = parser::ModuleIdent::new(name);
            parser.add_module(&module, &content);
            let parsed = parser.parse();

            let proto = parsed.generate_module_proto(&module);
            let mut buf = Vec::new();
            proto.encode(&mut buf).unwrap();

            let mut file = File::create(output).unwrap();
            file.write_all(&buf).unwrap();
        }
        Some(Commands::DumpModuleSchema { file }) => {
            info!("Dumping {:?}", file);
            let reader = std::fs::File::open(&file).expect("unable to open file");
            let module = ftl::schema::binary_to_module(reader);
            serde_json::to_writer_pretty(std::io::stdout(), &module).unwrap();
        }
        None => {
            error!("No command given");
        }
    }
}
