use proc_macro::TokenStream;

/// ftl::verb is a "tag" only proc macro that does not generate any code, just keeps the original.
#[proc_macro_attribute]
pub fn verb(attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}
