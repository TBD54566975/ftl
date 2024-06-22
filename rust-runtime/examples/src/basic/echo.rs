use std::error::Error;

use serde::{Deserialize, Serialize};

use ftl::Context;

#[derive(Debug, Serialize, Deserialize)]
struct Request {
    pub name: String,
    pub age: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct Response {
    pub message: String,
}

// pub async fn test_verb(ctx: &Context, request: Request) -> Result<Response, Box<dyn Error>> {
#[ftl::verb]
pub async fn test_verb(ctx: Context, request: ()) -> Result<(), Box<dyn Error>> {
    println!("test_verb was called!");
    // let response = ctx.call(module::other_verb, request).await?;

    // Ok(Response {
    //     message: "Hello, World!".to_string(),
    // })

    Ok(())
}
