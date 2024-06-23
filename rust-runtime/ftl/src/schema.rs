//! A crate for parsing/generating code to and from schema binary.
use proc_macro2::Ident;
use prost::Message;
use syn::Path;

use ftl_protos::schema;

pub fn binary_to_module(mut reader: impl std::io::Read) -> schema::Module {
    let mut buf = Vec::new();
    reader.read_to_end(&mut buf).unwrap();
    schema::Module::decode(&buf[..]).unwrap()
}

pub fn code_to_module(module: &Ident, code: &str) -> schema::Module {
    let ast = syn::parse_file(code).unwrap();
    let verbs = extract_ast_verbs(module, ast);

    let verbs = verbs.iter().filter_map(|verb| {
        let has_ctx = fn_has_context_as_first_arg(&verb.func);
        if !has_ctx {
            panic!("First argument must be of type Context");
        }

        Some(verb.to_proto())
    });

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

#[derive(Debug)]
pub struct VerbToken {
    pub module: Ident,
    pub func: syn::ItemFn,
}

impl VerbToken {
    fn to_proto(&self) -> schema::Verb {
        let mut verb = schema::Verb::default();
        verb.name = self.func.sig.ident.to_string();

        let syn::FnArg::Typed(arg) = self.func.sig.inputs.first().unwrap() else {
            panic!("Function must have at least one argument");
        };

        let syn::Type::Reference(path) = &*arg.ty else {
            panic!("First argument must not be a self argument");
        };

        let syn::Type::Path(type_path) = &*path.elem else {
            panic!("First argument must be of type Path");
        };

        let context_path = syn::parse_str("Context").unwrap();
        if type_path.path != context_path {
            panic!("First argument must be of type Context");
        }

        schema::Verb {
            runtime: None,
            pos: None,
            comments: vec![],
            export: false,
            name: self.func.sig.ident.to_string(),
            request: None,
            response: None,
            metadata: vec![],
        }
    }

    pub fn get_request_type(&self) -> Box<syn::Type> {
        let syn::FnArg::Typed(arg) = self.func.sig.inputs.last().unwrap() else {
            panic!("Function must have two arguments");
        };

        arg.ty.clone()
    }
}

/// Extract functions that are annotated with #[ftl::verb] and extract the AST node.
pub fn extract_ast_verbs(module: &Ident, ast: syn::File) -> Vec<VerbToken> {
    let ftl_verb_path = syn::parse_str("ftl::verb").unwrap();

    let mut verbs = vec![];
    for item in ast.items {
        let item = match item {
            syn::Item::Fn(func) => func,
            _ => continue,
        };

        if !has_meta_path(&item.attrs, &ftl_verb_path) {
            continue;
        }

        verbs.push(VerbToken {
            module: module.clone(),
            func: item,
        });
    }
    verbs
}

// Look for #[path_str] e.g. #[ftl::verb], and extract the function signature.
fn has_meta_path(attrs: &[syn::Attribute], expected_path: &Path) -> bool {
    attrs.iter().any(|attr| attr.meta.path() == expected_path)
}

// TODO: make this a bit less overly specific. eg require_arg_type(&func, 0, "Context")
fn fn_has_context_as_first_arg(func: &syn::ItemFn) -> bool {
    let Some(arg) = func.sig.inputs.first() else {
        println!("Function must have at least one argument");
        return false;
    };

    let syn::FnArg::Typed(pat) = arg else {
        println!("First argument must not be a self argument");
        return false;
    };

    let syn::Type::Reference(path) = &*pat.ty else {
        println!(
            "First argument must be of type Reference instead of {:?}",
            pat.ty
        );
        return false;
    };
    let syn::Type::Path(type_path) = &*path.elem else {
        println!(
            "First argument must be of type Path instead of {:?}",
            path.elem
        );
        return false;
    };

    let context_path = syn::parse_str("Context").unwrap();
    if type_path.path != context_path {
        println!(
            "First argument must be of type Context instead of {:?}",
            type_path.path
        );
        return false;
    }

    true
}

#[cfg(test)]
mod tests {
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

        let m = code_to_module(code);
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
