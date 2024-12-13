use std::process::Command;
use std::path::PathBuf;
use std::fs;
use prost::Message;
use tempfile::TempDir;
use sha2::{Sha256, Digest};

#[path = "../src/plugin/mod.rs"]
mod plugin;
#[path = "../src/protos/mod.rs"]
mod protos;

use protos::schemapb;

fn build_wasm() -> Result<(), Box<dyn std::error::Error>> {
    let status = Command::new("just")
        .arg("build-sqlc-gen-ftl")
        .status()?;

    if !status.success() {
        return Err("Failed to build WASM".into());
    }
    Ok(())
}

fn expected_module_schema() -> schemapb::Module {
    schemapb::Module {
        name: "echo".to_string(),
        builtin: false,
        runtime: None,
        comments: vec![],
        pos: None,
        decls: vec![
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "GetUserByIDQuery".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: vec![
                        schemapb::Field {
                            name: "id".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::Int(schemapb::Int {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        }
                    ],
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "GetUserByIDResult".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: vec![
                        schemapb::Field {
                            name: "id".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::Int(schemapb::Int {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        },
                        schemapb::Field {
                            name: "name".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::String(schemapb::String {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        },
                        schemapb::Field {
                            name: "email".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::String(schemapb::String {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        }
                    ],
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
                    name: "GetUserByID".to_string(),
                    export: false,
                    runtime: None,
                    request: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "GetUserByIDQuery".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    response: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "GetUserByIDResult".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Data(schemapb::Data {
                    name: "CreateUserQuery".to_string(),
                    export: false,
                    type_parameters: vec![],
                    fields: vec![
                        schemapb::Field {
                            name: "name".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::String(schemapb::String {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        },
                        schemapb::Field {
                            name: "email".to_string(),
                            r#type: Some(schemapb::Type {
                                value: Some(schemapb::r#type::Value::String(schemapb::String {
                                    pos: None,
                                }))
                            }),
                            pos: None,
                            comments: vec![],
                            metadata: vec![],
                        }
                    ],
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
            schemapb::Decl {
                value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
                    name: "CreateUser".to_string(),
                    export: false,
                    runtime: None,
                    request: Some(schemapb::Type {
                        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
                            module: "echo".to_string(),
                            name: "CreateUserQuery".to_string(),
                            pos: None,
                            type_parameters: vec![],
                        }))
                    }),
                    response: None,
                    pos: None,
                    comments: vec![],
                    metadata: vec![],
                })),
            },
        ],
    }
}

fn get_sqlc_config(wasm_path: &PathBuf) -> Result<String, Box<dyn std::error::Error>> {
    let wasm_contents = fs::read(wasm_path)?;
    let mut hasher = Sha256::new();
    hasher.update(&wasm_contents);
    let sha256_hash = hex::encode(hasher.finalize());

    Ok(format!(
        r#"version: '2'
plugins:
- name: ftl
  wasm:
    url: file://sqlc-gen-ftl.wasm
    sha256: {}
sql:
- schema: schema.sql
  queries: queries.sql
  engine: postgresql
  codegen:
  - out: gen
    plugin: ftl
    options:
      module: echo"#,
      sha256_hash,
    ))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    #[cfg_attr(not(feature = "ci"), ignore)]
    fn test_wasm_generate() -> Result<(), Box<dyn std::error::Error>> {
        if let Err(e) = build_wasm() {
            return Err(format!("Failed to build WASM: {}", e).into());
        }

        let temp_dir = TempDir::new()?;
        let gen_dir = temp_dir.path().join("gen");
        std::fs::create_dir(&gen_dir)?;
        
        let root_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
        let test_dir = root_dir.join("test");
        let wasm_path = temp_dir.path().join("sqlc-gen-ftl.wasm");

        std::fs::copy(
            test_dir.join("testdata/schema.sql"),
            temp_dir.path().join("schema.sql")
        )?;
        std::fs::copy(
            test_dir.join("testdata/queries.sql"),
            temp_dir.path().join("queries.sql")
        )?;
        std::fs::copy(
            root_dir.join("target/wasm32-wasip1/release/sqlc-gen-ftl.wasm"),
            &wasm_path
        )?;
        
        let config_contents = get_sqlc_config(&wasm_path)?;
        let config_path = temp_dir.path().join("sqlc.yaml");
        std::fs::write(&config_path, config_contents)?;

        let output = Command::new("sqlc")
            .arg("generate")
            .arg("--file")
            .arg(&config_path)
            .current_dir(temp_dir.path())
            .env("SQLC_VERSION", "dev")
            .env("SQLCDEBUG", "true")
            .output()?;

        if !output.status.success() {
            return Err(format!(
                "sqlc generate failed with status: {}\nstderr: {}",
                output.status,
                String::from_utf8_lossy(&output.stderr)
            ).into());
        }

        let pb_contents = std::fs::read(gen_dir.join("queries.pb"))?;
        let actual_module = schemapb::Module::decode(&*pb_contents)?;
        let expected_module = expected_module_schema();

        assert_eq!(
            &actual_module, 
            &expected_module, 
            "Schema mismatch.\nActual: {:#?}\nExpected: {:#?}",
            actual_module,
            expected_module
        );

        Ok(())
    }
}

