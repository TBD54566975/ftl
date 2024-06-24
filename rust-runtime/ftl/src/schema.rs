//! A crate for parsing/generating code to and from schema binary.
use prost::Message;

use ftl_protos::schema;

use crate::parser::{ModuleIdent, Parser};

pub fn binary_to_module(mut reader: impl std::io::Read) -> schema::Module {
    let mut buf = Vec::new();
    reader.read_to_end(&mut buf).unwrap();
    schema::Module::decode(&buf[..]).unwrap()
}

impl Parser {
    pub fn generate_module_proto(&self, module: &ModuleIdent) -> schema::Module {
        let verbs = &self.verb_tokens.get(module).unwrap();

        let verbs = verbs.iter().map(|verb| verb.to_proto());

        let mut decls = vec![];
        decls.extend(verbs.into_iter().map(|verb| schema::Decl {
            value: Some(schema::decl::Value::Verb(verb)),
        }));

        schema::Module {
            runtime: None,
            pos: None,
            comments: vec![],
            builtin: false,
            name: "".to_string(),
            decls,
        }
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
        pub async fn test_verb(ctx: &Context, request: Request) -> Result<Response, Box<dyn Error>> {
            // let response = ctx.call(module::other_verb, request).await?;

            Ok(Response {
                message: format!("Hello {}!", request.name),
            })
        }
        "#;

        let mut parser = Parser::new();
        let moo = ModuleIdent::new("moo");
        parser.add_module(&moo, code);
        let m = parser.generate_module_proto(&moo);

        assert_eq!(
            m,
            schema::Module {
                runtime: None,
                pos: None,
                comments: vec![],
                builtin: false,
                name: "".to_string(),
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
