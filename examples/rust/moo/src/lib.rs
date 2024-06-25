use std::error::Error;
use ftl::Context;

pub struct Request {
    pub name: String,
    pub age: u32,
}

pub struct Response {
    pub message: String,
}

#[ftl::verb]
pub async fn test_verb(ctx: &Context, request: Request) -> Result<Response, Box<dyn Error>> {
    // let response = ctx.call(module::other_verb, request).await?;

    Ok(Response {
        message: format!("Hello {}!", request.name),
    })
}
