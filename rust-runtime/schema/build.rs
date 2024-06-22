use std::env;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // println!("cargo:rustc-env=PROTOC=../../../bin/protoc"); // TODO: Hacks!

    // tonic_build::configure()
    //     // .type_attribute(".", "#[derive(serde::Serialize)]")
    //     .compile(
    //         &["../../backend/protos/xyz/block/ftl/v1/schema/schema.proto"],
    //         &["../../backend/protos"],
    //     )?;
    // Ok(())

    // let root = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("protos");
    // let proto_files = vec![root.join("../protos/xyz/block/ftl/v1/schema/schema.proto")];

    let root = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../../backend/protos");
    let proto_files = vec![root.join("xyz/block/ftl/v1/schema/schema.proto")];

    // Tell cargo to recompile if any of these proto files are changed
    for proto_file in &proto_files {
        println!("cargo:rerun-if-changed={}", proto_file.display());
    }

    let descriptor_path = PathBuf::from(env::var("OUT_DIR").unwrap()).join("proto_descriptor.bin");

    prost_build::Config::new()
        // Save descriptors to file
        .file_descriptor_set_path(&descriptor_path)
        // Override prost-types with pbjson-types
        .compile_well_known_types()
        .extern_path(".google.protobuf", "::pbjson_types")
        // Generate prost structs
        .compile_protos(&proto_files, &[root])?;

    let descriptor_set = std::fs::read(descriptor_path)?;
    pbjson_build::Builder::new()
        .register_descriptors(&descriptor_set)?
        .build(&[".xyz.block.ftl.v1.schema"])?;

    Ok(())
}
