use std::error::Error;

use serde::{Deserialize, Serialize};
use tracing::info;

use ftl::Context;

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
pub async fn unit_to_unit(ctx: Context, request: ()) -> Result<(), Box<dyn Error>> {
    info!("unit_to_unit was called!");
    Ok(())
}

#[ftl::verb]
pub async fn request_to_unit(ctx: Context, request: Request) -> Result<(), Box<dyn Error>> {
    info!(?request, "request_to_unit was called!");

    Ok(())
}
