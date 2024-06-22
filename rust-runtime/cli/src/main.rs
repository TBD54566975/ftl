use std::path::PathBuf;

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    Serve {
        #[arg(short, long, default_value = "8080")]
        port: u16,
    },
    DumpModule {
        file: PathBuf,
    },
}

fn main() {
    let cli = Cli::parse();
    tracing_subscriber::fmt::init();

    match cli.command {
        Some(Commands::Serve { port }) => {
            eprintln!("Serving on port {}", port);
        }
        Some(Commands::DumpModule { file }) => {
            eprintln!("Dumping {:?}", file);
            let reader = std::fs::File::open(&file).expect("unable to open file");
            let module = schema::binary_to_module(reader);
            serde_json::to_writer_pretty(std::io::stdout(), &module).unwrap();
        }
        None => {
            eprintln!("No command given");
        }
    }
}
