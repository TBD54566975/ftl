mod protos;
mod plugin;

use std::io::{Read, Write};
pub use plugin::Plugin;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut input_bytes = Vec::new();
    if let Err(e) = std::io::stdin().read_to_end(&mut input_bytes) {
        eprintln!("Failed to read stdin: {}", e);
        std::process::exit(1);
    }

    match Plugin::generate_from_input(&input_bytes) {
        Ok(output_bytes) => {
            if let Err(e) = std::io::stdout().write_all(&output_bytes.as_slice()) {
                eprintln!("Failed to write to stdout: {}", e);
                std::process::exit(1);
            }
            Ok(())
        }
        Err(e) => {
            eprintln!("Plugin execution failed: {}", e);
            std::process::exit(1);
        }
    }
}
