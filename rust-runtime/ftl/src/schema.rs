//! A crate for parsing/generating code to and from schema binary.
use prost::Message;

use ftl_protos::schema;

use crate::parser::{ident_from_type, ModuleIdent, Parsed, Parser, VerbToken};

pub fn binary_to_module(mut reader: impl std::io::Read) -> schema::Module {
    let mut buf = Vec::new();
    reader.read_to_end(&mut buf).unwrap();
    schema::Module::decode(&buf[..]).unwrap()
}

impl Parsed {
    pub fn generate_module_proto(&self, module: &ModuleIdent) -> schema::Module {
        let mut decls = vec![];

        for verb_token in &self.verbs {
            decls.push(self.get_decl_for_verb(verb_token));
        }
        for ident in self.types.keys() {
            decls.push(self.get_decl_for_data(ident));
        }

        dbg!(&decls);

        schema::Module {
            runtime: None,
            pos: None,
            comments: vec![],
            builtin: false,
            name: module.0.to_string(),
            decls,
        }
    }

    fn get_decl_for_verb(&self, verb_token: &VerbToken) -> schema::Decl {
        let verb = verb_token.to_verb_proto();
        schema::Decl {
            value: Some(schema::decl::Value::Verb(verb)),
        }
    }

    fn get_decl_for_data(&self, ident: &syn::Ident) -> schema::Decl {
        let item = self.types.get(ident).unwrap();
        match item {
            syn::Item::Struct(item_struct) => {
                let mut fields = vec![];

                for field in &item_struct.fields {
                    // let ident = field.ty;
                    let type_ident = ident_from_type(&field.ty);
                    fields.push(schema::Field {
                        pos: None,
                        comments: vec![],
                        name: field.ident.as_ref().unwrap().to_string(),
                        r#type: Some(self.get_type_recursive(&type_ident)),
                        metadata: vec![],
                    });
                }

                let data = schema::Data {
                    pos: None,
                    comments: vec![],
                    export: true,
                    name: ident.to_string(),
                    type_parameters: vec![],
                    fields,
                    metadata: vec![],
                };

                schema::Decl {
                    value: Some(schema::decl::Value::Data(data)),
                }
            }
            _ => todo!(),
        }
    }

    fn get_type_recursive(&self, ident: &syn::Ident) -> schema::Type {
        let maybe_value = match ident.to_string().as_str() {
            "String" => Some(schema::r#type::Value::String(schema::String { pos: None })),
            "u32" => Some(schema::r#type::Value::Int(schema::Int { pos: None })),
            _ => None,
        };
        if let Some(value) = maybe_value {
            return schema::Type { value: Some(value) };
        }

        println!("maybe not an internal type: {:?}", ident);

        let item = self
            .types
            .get(ident)
            .expect(format!("type not found: {:?}", ident).as_str());
        let value = match item {
            syn::Item::Struct(item_struct) => schema::r#type::Value::Ref(schema::Ref {
                pos: None,
                name: ident.to_string(),
                module: "".to_string(),
                type_parameters: vec![],
            }),
            unknown => todo!("unhandled type {:?}", unknown),
            // Item::Const(_) => {}
            // Item::Enum(_) => {}
            // Item::ExternCrate(_) => {}
            // Item::Fn(_) => {}
            // Item::ForeignMod(_) => {}
            // Item::Impl(_) => {}
            // Item::Macro(_) => {}
            // Item::Mod(_) => {}
            // Item::Static(_) => {}
            // Item::Trait(_) => {}
            // Item::TraitAlias(_) => {}
            // Item::Type(_) => {}
            // Item::Union(_) => {}
            // Item::Use(_) => {}
            // Item::Verbatim(_) => {}
        };

        schema::Type { value: Some(value) }
    }
}

#[cfg(test)]
mod tests {
    use crate::parser::Parser;

    use super::*;

    #[test]
    fn it_works() {
        let config = schema::Config {
            pos: None,
            comments: vec![],
            name: "sup".to_string(),
            r#type: None,
        };

        let mut encoded = Vec::new();
        Message::encode(&config, &mut encoded).unwrap();

        let decoded = schema::Config::decode(&encoded[..]).unwrap();
        assert_eq!(config, decoded);

        dbg!(encoded);
        dbg!(decoded);
    }

    #[test]
    fn ast_to_proto() {
        //
        let code = r#"
        use ftl::Context;

        struct Request {
            pub name: String,
            pub age: u32,
        }

        struct Response {
            pub message: String,
        }

        #[ftl::verb]
        pub async fn test_verb(ctx: Context, request: Request) -> Result<Response, Box<dyn Error>> {
            // let response = ctx.call(module::other_verb, request).await?;

            Ok(Response {
                message: format!("Hello {}!", request.name),
            })
        }
        "#;

        let mut parser = Parser::new();
        let moo = ModuleIdent::new("moo");
        parser.add_module(&moo, code);
        let parsed = parser.parse();
        let m = parsed.generate_module_proto(&moo);

        assert_eq!(
            m,
            schema::Module {
                runtime: None,
                pos: None,
                comments: vec![],
                builtin: false,
                name: "moo".to_string(),
                decls: vec![schema::Decl {
                    value: Some(schema::decl::Value::Verb(schema::Verb {
                        runtime: None,
                        pos: None,
                        comments: vec![],
                        export: false,
                        name: "test_verb".to_string(),
                        request: None,
                        response: None,
                        metadata: vec![],
                    })),
                },],
            }
        );

        dbg!(m);
    }
}
