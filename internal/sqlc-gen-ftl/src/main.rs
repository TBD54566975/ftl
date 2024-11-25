mod protos;
mod plugin;

use std::io::{Read, Write};
pub use plugin::Plugin;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut input_bytes = Vec::new();
    if std::io::stdin().read_to_end(&mut input_bytes).is_err() {
        std::process::exit(1);
    }

    match Plugin::generate_from_input(&input_bytes) {
        Ok(output_bytes) => {
            if std::io::stdout().write_all(&output_bytes.as_slice()).is_err() {
                std::process::exit(1);
            }
            Ok(())
        }
        Err(_) => {
            std::process::exit(1);
        }
    }
}
