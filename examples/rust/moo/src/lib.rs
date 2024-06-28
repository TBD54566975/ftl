use ftl::Context;
use serde::{Deserialize, Serialize};
use tracing::info;
use std::collections::HashMap;
use std::fmt::Debug;

mod builtin {
    use super::*;

    #[derive(Debug, Serialize, Deserialize)]
    pub struct HttpRequest<Body>
    where Body: Debug
    {
        pub method: String,
        pub path: String,
        pub path_parameters: HashMap<String, String>,
        pub query: HashMap<String, Vec<String>>,
        pub headers: HashMap<String, Vec<String>>,
        pub body: Body,
    }

    #[derive(Debug, Serialize, Deserialize)]
    pub struct HttpResponse<Body, Error>
    where
        Body: Debug,
        Error: Debug
    {
        pub status: i32,
        pub headers: HashMap<String, Vec<String>>,
        pub body: Option<Body>,
        pub error: Option<Error>,
    }

    #[derive(Debug, Serialize, Deserialize)]
    pub struct Empty {}
}

mod echo {
    use super::*;

    #[derive(Debug, Serialize, Deserialize)]
    pub struct EchoRequest {
        pub name: String,
    }

    #[derive(Debug, Serialize, Deserialize)]
    pub struct EchoResponse {
        pub message: String,
    }

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
pub struct MooVerbRequest {
    pub your_name: String,
    pub age: u32,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MooVerbResponse {
    pub message: String,
    pub your_score: u32,
}

#[ftl::verb]
pub async fn test_verb(mut ctx: Context, request: MooVerbRequest) -> MooVerbResponse {
    info!("test_verb was called");
    info!("request: {:?}", &request);

    let echo_response = ctx.call(echo::EchoVerb, echo::EchoRequest {
        name: request.your_name,
    }).await;

    MooVerbResponse {
        message: format!("Hello. Ping response was: {}!", echo_response.message),
        your_score: request.age * 42,
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IngressRequest {
    pub user_name: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IngressResponse {
    pub message: String,
}

#[ftl::ingress]
pub async fn http_echo(_: Context, request: builtin::HttpRequest<IngressRequest>) -> builtin::HttpResponse<IngressResponse, ()> {
    info!("http_echo was called");
    info!("request: {:?}", &request);

    builtin::HttpResponse {
        status: 200,
        headers: HashMap::new(),
        body: Some(IngressResponse {
            message: format!("Hello, {}!", request.body.user_name),
        }),
        error: None,
    }
}