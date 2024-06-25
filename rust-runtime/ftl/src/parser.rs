use std::collections::HashMap;
use std::path::{Path, PathBuf};

use proc_macro2::{Ident, Span};

use ftl_protos::schema;

pub struct Parser {
    pub verb_tokens: HashMap<ModuleIdent, Vec<VerbToken>>,
}

impl Parser {
    pub fn new() -> Self {
        Self {
            verb_tokens: HashMap::new(),
        }
    }

    pub fn add_module(&mut self, module: &ModuleIdent, code: &str) {
        let ast = syn::parse_file(code).unwrap();
        let verbs = extract_ast_verbs(module, ast);

        self.verb_tokens.insert(module.clone(), verbs);
    }

    pub fn add_glob(&mut self, path: &Path) {
        let sources = find_sources(path);
        for source in sources {
            let contents = std::fs::read_to_string(&source).unwrap();
            let file_name = source.file_stem().unwrap().to_str().unwrap();
            let mod_ident = ModuleIdent::new(file_name);

            self.add_module(&mod_ident, &contents);
        }
    }

    pub fn modules_count(&self) -> usize {
        self.verb_tokens.len()
    }
}

/// Find all .rs sources in src.
fn find_sources(path: &Path) -> Vec<PathBuf> {
    glob::glob(path.to_str().unwrap())
        .unwrap()
        .map(|entry| entry.unwrap())
        .map(|entry| entry.as_path().to_path_buf())
        .collect()
}

#[derive(Debug, Clone, Hash, Eq, PartialEq)]
pub struct ModuleIdent(pub Ident);

impl ModuleIdent {
    pub fn new(name: &str) -> Self {
        Self(Ident::new(name, Span::call_site()))
    }
}

#[derive(Debug)]
pub struct VerbToken {
    pub module: ModuleIdent,
    pub func: syn::ItemFn,
}

impl VerbToken {
    pub fn try_parse_any_item(module: &ModuleIdent, item: syn::Item) -> Option<Self> {
        let func = match item {
            syn::Item::Fn(func) => func,
            // Quietly ignore non-functions.
            _ => return None,
        };

        let ftl_verb_path = syn::parse_str("ftl::verb").unwrap();
        if !has_meta_path(&func.attrs, &ftl_verb_path) {
            // No #[ftl::verb] annotation, ignore.
            return None;
        }

        VerbToken::ensure_fn_has_context_as_first_arg(&func);

        Some(VerbToken {
            module: module.clone(),
            func,
        })
    }

    pub fn to_proto(&self) -> schema::Verb {
        let mut verb = schema::Verb::default();
        verb.name = self.func.sig.ident.to_string();

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

    // TODO: make this a bit less overly specific. eg require_arg_type(&func, 0, "Context")
    pub fn ensure_fn_has_context_as_first_arg(func: &syn::ItemFn) {
        let Some(arg) = func.sig.inputs.first() else {
            panic!("Function must have at least one argument");
        };

        let syn::FnArg::Typed(pat) = arg else {
            dbg!(arg);
            panic!("First argument must not be a self argument");
        };

        let syn::Type::Path(path) = &*pat.ty else {
            panic!(
                "First argument must be of type Path instead of {:?}",
                pat.ty
            );
        };
        let path = &path.path else {
            panic!(
                "First argument must be of type Path instead of {:?}",
                path.path
            );
        };

        // let context_path = syn::parse_str("Context").unwrap();
        // if path != context_path {
        //     println!(
        //         "First argument must be of type Context instead of {:?}",
        //         path
        //     );
        //     return false;
        // }

        dbg!(path);
    }
}

/// Extract functions that are annotated with #[ftl::verb] and extract the AST node.
pub fn extract_ast_verbs(module: &ModuleIdent, ast: syn::File) -> Vec<VerbToken> {
    let mut verbs = vec![];
    for item in ast.items {
        let Some(verb_token) = VerbToken::try_parse_any_item(&module, item) else {
            continue;
        };

        verbs.push(verb_token);
    }
    verbs
}

// Look for #[path_str] e.g. #[ftl::verb], and extract the function signature.
fn has_meta_path(attrs: &[syn::Attribute], expected_path: &syn::Path) -> bool {
    attrs.iter().any(|attr| attr.meta.path() == expected_path)
}
