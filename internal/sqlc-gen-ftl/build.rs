fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut config = prost_build::Config::new();
    config
        .out_dir("src/protos")
        .compile_protos(
            &["proto/codegen.proto"],
            &["proto/"]
        )?;
    Ok(())
}
