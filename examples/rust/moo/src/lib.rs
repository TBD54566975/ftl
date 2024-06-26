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
}

#[ftl::verb]
pub async fn test_verb(ctx: Context, request: Request) -> Result<Response, Box<dyn Error>> {
    info!("test_verb was called");
    // let response = ctx.call(module::other_verb, request).await?;

    Ok(Response {
        message: format!("Hello {}!", request.name),
    })
}
