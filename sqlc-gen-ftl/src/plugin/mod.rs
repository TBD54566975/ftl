#![allow(dead_code)]

use crate::protos::pluginpb;
use crate::protos::schemapb;
use crate::protos::schemapb::r#type::Value as TypeValue;
use prost::Message;
use std::io;

pub struct Plugin;

impl Plugin {
    pub fn generate_from_input(input: &[u8]) -> Result<Vec<u8>, io::Error> {
        let req = pluginpb::GenerateRequest::decode(input)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        let resp = Self::handle_generate(req)?;
        Ok(resp.encode_to_vec())
    }

    fn handle_generate(req: pluginpb::GenerateRequest) -> Result<pluginpb::GenerateResponse, io::Error> {
        let module = generate_schema(&req)?;
        Ok(pluginpb::GenerateResponse {
            files: vec![pluginpb::File {
                name: "queries.pb".to_string(),
                contents: module.encode_to_vec(),
            }],
        })
    }
}

fn generate_schema(request: &pluginpb::GenerateRequest) -> Result<schemapb::Module, io::Error> {
    let mut decls = Vec::new();
    let module_name = get_module_name(request)?;
    
    for query in &request.queries {
        if !query.params.is_empty() {
            decls.push(to_verb_request(query));
        }

        if !query.columns.is_empty() {
            decls.push(to_verb_response(query));
        }

        decls.push(to_verb(query, &module_name));
    }

    Ok(schemapb::Module {
        name: module_name,
        builtin: false,
        runtime: None,
        comments: Vec::new(),
        pos: None,
        decls,
    })
}

fn to_verb(query: &pluginpb::Query, module_name: &str) -> schemapb::Decl {
    let request_type = if !query.params.is_empty() {
        Some(to_schema_ref(module_name, &format!("{}Query", query.name)))
    } else {
        None
    };

    let response_type = if query.cmd == ":exec" {
        None
    } else {
        Some(to_schema_ref(module_name, &format!("{}Result", query.name)))
    };

    schemapb::Decl {
        value: Some(schemapb::decl::Value::Verb(schemapb::Verb {
            name: query.name.clone(),
            export: false,
            runtime: None,
            request: request_type,
            response: response_type,
            pos: None,
            comments: Vec::new(),
            metadata: Vec::new(),
        })),
    }
}

fn to_verb_request(query: &pluginpb::Query) -> schemapb::Decl {
    schemapb::Decl {
        value: Some(schemapb::decl::Value::Data(schemapb::Data {
            name: format!("{}Query", query.name),
            export: false,
            type_parameters: Vec::new(),
            fields: query.params.iter().map(|param| {
                let name = param.column.as_ref()
                    .map(|col| col.name.clone())
                    .unwrap_or_else(|| format!("param{}", param.number));
                let sql_type = param.column.as_ref().and_then(|col| col.r#type.as_ref());
                to_schema_field(name, sql_type)
            }).collect(),
            pos: None,
            comments: Vec::new(),
            metadata: Vec::new(),
        })),
    }
}

fn to_verb_response(query: &pluginpb::Query) -> schemapb::Decl {
    schemapb::Decl {
        value: Some(schemapb::decl::Value::Data(schemapb::Data {
            name: format!("{}Result", query.name),
            export: false,
            type_parameters: Vec::new(),
            fields: query.columns.iter().map(|col| {
                to_schema_field(col.name.clone(), col.r#type.as_ref())
            }).collect(),
            pos: None,
            comments: Vec::new(),
            metadata: Vec::new(),
        })),
    }
}

fn to_schema_field(name: String, sql_type: Option<&pluginpb::Identifier>) -> schemapb::Field {
    schemapb::Field {
        name,
        r#type: Some(sql_type.map_or_else(
            || schemapb::Type {
                value: Some(TypeValue::Any(schemapb::Any { pos: None })),
            },
            to_schema_type
        )),
        pos: None,
        comments: Vec::new(),
        metadata: Vec::new(),
    }
}

fn to_schema_ref(module_name: &str, name: &str) -> schemapb::Type {
    schemapb::Type {
        value: Some(schemapb::r#type::Value::Ref(schemapb::Ref {
            module: module_name.to_string(),
            name: name.to_string(),
            pos: None,
            type_parameters: vec![],
        }))
    }
}

fn to_schema_type(sql_type: &pluginpb::Identifier) -> schemapb::Type {
    let value = match sql_type.name.as_str() {
        "integer" | "bigint" | "smallint" | "serial" | "bigserial" => 
            TypeValue::Int(schemapb::Int { pos: None }),
        "real" | "float" | "double" | "numeric" | "decimal" => 
            TypeValue::Float(schemapb::Float { pos: None }),
        "text" | "varchar" | "char" | "uuid" => 
            TypeValue::String(schemapb::String { pos: None }),
        "boolean" => 
            TypeValue::Bool(schemapb::Bool { pos: None }),
        "timestamp" | "date" | "time" => 
            TypeValue::Time(schemapb::Time { pos: None }),
        "json" | "jsonb" => 
            TypeValue::Any(schemapb::Any { pos: None }),
        "bytea" | "blob" => 
            TypeValue::Bytes(schemapb::Bytes { pos: None }),
        _ => 
            TypeValue::Any(schemapb::Any { pos: None }),
    };
    
    schemapb::Type {
        value: Some(value),
    }
}

fn get_module_name(req: &pluginpb::GenerateRequest) -> Result<String, io::Error> {
    let codegen = req.settings
        .as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing settings"))?
        .codegen
        .as_ref()
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing codegen settings"))?;

    let options_str = String::from_utf8(codegen.options.clone())
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, format!("Invalid UTF-8 in options: {}", e)))?;
    
    let options: serde_json::Value = serde_json::from_str(&options_str)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, format!("Failed to parse JSON options: {}", e)))?;

    options.get("module")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
        .ok_or_else(|| io::Error::new(io::ErrorKind::InvalidData, "Missing module name in options"))
}
