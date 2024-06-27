use std::error::Error;
use ftl::Context;
use serde::{Deserialize, Serialize};
use tracing::info;

#[derive(Debug, Serialize, Deserialize)]
pub struct Request {
    pub name: String,
    pub age: u32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Response {
    pub message: String,
    pub score: u32,
}

#[ftl::verb]
pub async fn test_verb(ctx: Context, request: Request) -> Response {
    info!("test_verb was called");
    info!("request: {:?}", request);
    // let response = ctx.call(module::other_verb, request).await?;

    Response {
        message: format!("Hello {}!", request.name),
        score: request.age * 42,
    }
}
