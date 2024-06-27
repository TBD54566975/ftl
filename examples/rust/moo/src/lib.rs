use std::error::Error;
use ftl::Context;
use serde::{Deserialize, Serialize};
use tracing::info;
// use echo;

mod echo {
    #[derive(Debug, ::serde::Serialize, ::serde::Deserialize)]
    pub struct EchoRequest {
        pub name: String,
    }

    #[derive(Debug, ::serde::Serialize, ::serde::Deserialize)]
    pub struct EchoResponse {
        pub message: String,
    }

    // scaffolding
    // pub fn echo(_ctx: ftl::Context, _request: EchoRequest) -> EchoResponse {
    //     panic!("Do not call this directly!")
    // }

    // impl VerbFn for echo {
    //     fn module_and_verb() -> (String, String) {
    //         ("echo".to_string(), "echo".to_string())
    //     }
    // }

    pub struct EchoVerb;

    impl ftl::VerbFn for EchoVerb {
        type Request = EchoRequest;
        type Response = EchoResponse;

        fn module() -> &'static str {
            "echo"
        }
        fn name() -> &'static str {
            "echo"
        }
    }

}

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
pub async fn test_verb(mut ctx: Context, request: Request) -> Response {
    info!("test_verb was called");
    info!("request: {:?}", &request);

    let echo_response = ctx.call(echo::EchoVerb, echo::EchoRequest {
        name: request.your_name,
    }).await;

    Response {
        message: format!("Hello. Ping response was: {}!", echo_response.message),
        your_score: request.age * 42,
    }
}

