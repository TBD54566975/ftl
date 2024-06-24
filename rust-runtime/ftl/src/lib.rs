pub use ftl_derive::verb;
pub use parser::build;

mod generator;
mod parser;
pub mod runner;
pub mod schema;
pub mod verb_server;

#[derive(Debug, Default)]
pub struct Context {}
