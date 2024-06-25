pub use ftl_derive::verb;

pub mod builder;
pub mod generator;
pub mod parser;
pub mod runner;
pub mod schema;
pub mod verb_server;

#[derive(Debug, Default)]
pub struct Context {}
