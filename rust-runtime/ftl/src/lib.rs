pub use build::build;
pub use ftl_derive::verb;
pub use server::serve;

mod build;
pub mod schema;
pub mod server;

#[derive(Debug, Default)]
pub struct Context {}
