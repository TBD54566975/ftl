use std::path::{Path, PathBuf};

use proc_macro2::{Ident, Span};

use ftl_protos::schema;

fn brainstorm() {
    let parsed = parse(Path::new("src").join("**/*.rs"));

    let out_dir = std::env::var("OUT_DIR").unwrap();
    let out_path = Path::new(&out_dir).join("call_immediate.rs");
    parsed.generate_call_immediate_file(&out_path);
    parsed.generate_protos();
}

pub fn parse(path: PathBuf) -> Parsed {
    let sources = find_sources();
    let mut verb_tokens = vec![];
    for source in sources {
        let contents = std::fs::read_to_string(&source).unwrap();
        let ast = syn::parse_file(&contents).unwrap();
        let file_name = source.file_stem().unwrap().to_str().unwrap();
        let mod_ident = Ident::new(file_name, Span::call_site());
        verb_tokens.extend(extract_ast_verbs(&mod_ident, ast));
    }

    Parsed { verb_tokens }
}

pub struct Parsed {
    pub verb_tokens: Vec<VerbToken>,
}

/// Find all .rs sources in src.
fn find_sources(path: &Path) -> Vec<PathBuf> {
    glob::glob(path.to_str().unwrap())
        .unwrap()
        .map(|entry| entry.unwrap())
        .map(|entry| entry.as_path().to_path_buf())
        .collect()
}

#[derive(Debug)]
pub struct VerbToken {
    pub module: Ident,
    pub func: syn::ItemFn,
}

impl VerbToken {
    pub(crate) fn to_proto(&self) -> schema::Verb {
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
fn has_meta_path(attrs: &[syn::Attribute], expected_path: &syn::Path) -> bool {
    attrs.iter().any(|attr| attr.meta.path() == expected_path)
}

// TODO: make this a bit less overly specific. eg require_arg_type(&func, 0, "Context")
pub fn fn_has_context_as_first_arg(func: &syn::ItemFn) -> bool {
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
