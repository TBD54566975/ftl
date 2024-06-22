use proc_macro::TokenStream;

/// ftl::verb is a "tag" only proc macro that does not generate any code, just keeps the original.
#[proc_macro_attribute]
pub fn verb(attr: TokenStream, item: TokenStream) -> TokenStream {
    // // let input = parse_macro_input!(item as ItemFn);
    //
    // // Generate the output tokens as a TokenStream
    // let output = quote! {
    //     #item
    // };
    //
    // output.into()
    item
}
