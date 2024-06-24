use std::path::PathBuf;

use clap::{Parser, Subcommand};
use tracing::{error, info};

#[derive(Parser)]
#[command(version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    CallVerb {
        module: String,
        verb: String,
        request: String,
    },
    DumpModule {
        file: PathBuf,
    },
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    tracing_subscriber::fmt::init();

    match cli.command {
        Some(Commands::CallVerb {
            module,
            verb,
            request,
        }) => {
            info!("Calling verb {} in module {}", verb, module);
            // ftl::verb_client::call_verb(module, verb, request).await;
        }
        Some(Commands::DumpModule { file }) => {
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
