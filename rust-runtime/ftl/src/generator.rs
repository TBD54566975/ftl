use std::path::Path;

use quote::quote;

use crate::parser::{Parsed, VerbToken};

impl Parsed {
    pub fn generate_call_immediate_file(&self, out_path: &Path) {
        let mut call_immediate_case_tokens = vec![];
        for verb_token in &self.verbs {
            call_immediate_case_tokens.push(Self::to_call_immediate_case_token(verb_token));
        }

        let token_stream = quote::quote! {
            // Output should be boxed trait of Deserializable
            pub fn call_immediate(ctx: ::ftl::Context, module: String, verb: String, request_body: String)
                -> ::std::pin::Pin<Box<dyn ::std::future::Future<Output = String> + Send>> {
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

        eprintln!("Generating to: {}", out_path.display());
        eprintln!("Generated: {}", formatted_code);
        std::fs::write(out_path, formatted_code).unwrap();
    }

    pub fn to_call_immediate_case_token(verb_token: &VerbToken) -> proc_macro2::TokenStream {
        let module_name = verb_token.module.0.clone();
        let verb_name = verb_token.func.sig.ident.clone();
        let module_name_str = module_name.to_string();
        let verb_name_str = verb_name.to_string();
        let request_type = &verb_token.request_ident;
        let response_type = &verb_token.response_ident;

        // request type only supports existing in the same module or unit
        dbg!(&verb_token.request_ident);
        // if verb_token.request_ident == syn::Ident::new("()", proc_macro2::Span::call_site()) {
        //     quote! {
        //         (#module_name_str, #verb_name_str) => {
        //             #module_name::#verb_name(ctx, ()).await.unwrap();
        //         }
        //     }
        // } else {
        quote! {
            (#module_name_str, #verb_name_str) => {
                let request = ::serde_json::from_str::<#module_name::#request_type>(&request_body).unwrap();
                let response: #module_name::#response_type = #module_name::#verb_name(ctx, request).await;
                serde_json::to_string(&response).unwrap()
            }
        }
        // }
    }
}
