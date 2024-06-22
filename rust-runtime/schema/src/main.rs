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
    DumpModule { file: PathBuf },
}

fn main() {
    let cli = Cli::parse();

    match cli.command {
        Some(Commands::DumpModule { file }) => {
            eprintln!("Dumping {:?}", file);
            let reader = std::fs::File::open(&file).expect("unable to open file");
            let module = schema::module_from_binary(reader);
            serde_json::to_writer_pretty(std::io::stdout(), &module).unwrap();
        }
        None => {
            eprintln!("No command given");
        }
    }
}
