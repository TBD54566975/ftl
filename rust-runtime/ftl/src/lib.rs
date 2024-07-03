use serde::{de::DeserializeOwned, Serialize};

pub use ftl_derive::verb;
use ftl_protos::ftl::call_response::Response;
use ftl_protos::ftl::CallRequest;
use ftl_protos::ftl::verb_service_client::VerbServiceClient;
use ftl_protos::schema::Ref;

pub mod builder;
pub mod generator;
pub mod parser;
pub mod runtime;
pub mod schema;
pub mod verb_server;

pub trait VerbFn {
    type Request: Serialize;
    type Response: DeserializeOwned;

    fn module() -> &'static str;
    fn name() -> &'static str;
}

#[derive(Clone, Debug)]
pub struct Context {
    verb_client: VerbServiceClient<tonic::transport::Channel>,
}

impl Context {
    pub fn new(verb_client: VerbServiceClient<tonic::transport::Channel>) -> Self {
        Self { verb_client }
    }

    pub async fn call<V>(&mut self, _v: V, request: V::Request) -> V::Response
    where
        V: VerbFn,
    {
        let name = V::name().to_string();
        let module = V::module().to_string();

        let response = self
            .verb_client
            .call(tonic::Request::new(CallRequest {
                metadata: None,
                verb: Some(Ref {
                    pos: None,
                    name,
                    module,
                    type_parameters: vec![],
                }),
                body: serde_json::to_vec(&request).unwrap(),
            }))
            .await
            .unwrap()
            .into_inner();

        let response = response.response.unwrap();
        match response {
            Response::Body(vec) => serde_json::from_slice(&vec).unwrap(),
            Response::Error(err) => {
                panic!("Error: {:?}", err)
            }
        }
    }
}
