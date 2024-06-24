use std::path::{Path, PathBuf};

use proc_macro2::{Ident, Span};
use quote::quote;

use crate::schema::{extract_ast_verbs, VerbToken};

/// This runs in build.rs
/// Find all .rs sources in src.
/// For each source file, run code_to_module.
pub fn build() {
    let sources = find_sources();
    let mut call_immediate_case_tokens = vec![];
    for source in sources {
        let contents = std::fs::read_to_string(&source).unwrap();
        let ast = syn::parse_file(&contents).unwrap();
        let file_name = source.file_stem().unwrap().to_str().unwrap();
        let mod_ident = Ident::new(file_name, Span::call_site());
        let tokens = extract_ast_verbs(&mod_ident, ast)
            .iter()
            .map(to_call_immediate_case_token)
            .collect::<Vec<_>>();

        call_immediate_case_tokens.extend(tokens);
    }

    let token_stream = quote::quote! {
        pub fn call_immediate(ctx: ::ftl::Context, module: String, verb: String, request_body: String) -> ::std::pin::Pin<Box<dyn ::std::future::Future<Output = ()> + Send + Sync>> {
             let fut = async move {
                match (module.as_str(), verb.as_str()) {
                    #(#call_immediate_case_tokens)*
                    unknown => panic!("Unknown verb: {:?}", unknown),
                }
             };

            Box::pin(fut)
        }
    };

    eprintln!("Generated: {}", token_stream);
    let formatted_code = prettyplease::unparse(&syn::parse2(token_stream).unwrap());

    let out_dir = std::env::var("OUT_DIR").unwrap();
    let out_path = Path::new(&out_dir).join("call_immediate.rs");
    eprintln!("Generating to: {}", out_path.display());
    eprintln!("Generated: {}", formatted_code);
    std::fs::write(out_path, formatted_code).unwrap();
}

pub fn to_call_immediate_case_token(verb_token: &VerbToken) -> proc_macro2::TokenStream {
    let module_name = verb_token.module.clone();
    let verb_name = verb_token.func.sig.ident.clone();
    let module_name_str = module_name.to_string();
    let verb_name_str = verb_name.to_string();
    let request_type = verb_token.get_request_type();

    // request type only supports existing in the same module or unit
    if matches!(*request_type, syn::Type::Tuple(_)) {
        quote! {
            (#module_name_str, #verb_name_str) => {
                #module_name::#verb_name(ctx, ()).await.unwrap();
            }
        }
    } else {
        quote! {
            (#module_name_str, #verb_name_str) => {
                let request = ::serde_json::from_str::<#module_name::#request_type>(&request_body).unwrap();
                #module_name::#verb_name(ctx, request).await.unwrap();
            }
        }
    }
}

/// Find all .rs sources in src.
fn find_sources() -> Vec<PathBuf> {
    let src = Path::new("src");
    glob::glob(src.join("**/*.rs").to_str().unwrap())
        .unwrap()
        .map(|entry| entry.unwrap())
        .map(|entry| entry.as_path().to_path_buf())
        .collect()
}
