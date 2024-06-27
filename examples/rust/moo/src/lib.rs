use std::error::Error;
use ftl::Context;
use serde::{Deserialize, Serialize};
use tracing::info;

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Request {
    pub your_name: String,
    pub age: u32,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Response {
    pub message: String,
    pub your_score: u32,
}

#[ftl::verb]
pub async fn test_verb(ctx: Context, request: Request) -> Response {
    info!("test_verb was called");
    info!("request: {:?}", request);
    // let response = ctx.call(module::other_verb, request).await?;

    Response {
        message: format!("Hello {}!", request.your_name),
        your_score: request.age * 42,
    }
}
