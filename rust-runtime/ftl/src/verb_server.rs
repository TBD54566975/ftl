use std::future::Future;
use std::pin::Pin;

use heck::ToSnakeCase;
use prost::Message;
use tonic::{Request, Response, Status, Streaming};
use tonic::codegen::tokio_stream::Stream;
use tonic::transport::Server;

use ftl_protos as protos;
use ftl_protos::ftl::verb_service_server::VerbService;

use crate::Context;

pub type CallImmediateReturn = Pin<Box<dyn Future<Output = String> + Send + Sync>>;

pub type CallImmediateFn = fn(Context, String, String, String) -> CallImmediateReturn;

#[derive(Debug)]
pub struct Config {
    pub bind: String,
    pub call_immediate: CallImmediateFn,
}

impl Config {
    pub fn bind_url(&self) -> String {
        format!("http://{}", self.bind)
    }
}

pub async fn serve(config: Config) -> () {
    let addr = config.bind.parse().unwrap();
    let service = FtlService { config };

    Server::builder()
        .add_service(protos::ftl::verb_service_server::VerbServiceServer::new(
            service,
        ))
        .serve(addr)
        .await
        .unwrap();

    ()
}

#[derive(Debug)]
pub struct FtlService {
    config: Config,
}

type ModuleContextResponseStream =
    Pin<Box<dyn Stream<Item = Result<protos::ftl::ModuleContextResponse, Status>> + Send>>;

type AcquireLeaseStream =
    Pin<Box<dyn Stream<Item = Result<protos::ftl::AcquireLeaseResponse, Status>> + Send>>;

#[tonic::async_trait]
impl VerbService for FtlService {
    async fn ping(
        &self,
        request: Request<protos::ftl::PingRequest>,
    ) -> Result<Response<protos::ftl::PingResponse>, Status> {
        Ok(Response::new(protos::ftl::PingResponse { not_ready: None }))
    }

    type GetModuleContextStream = ModuleContextResponseStream;

    async fn get_module_context(
        &self,
        request: Request<protos::ftl::ModuleContextRequest>,
    ) -> Result<Response<Self::GetModuleContextStream>, Status> {
        todo!()
    }

    type AcquireLeaseStream = AcquireLeaseStream;

    async fn acquire_lease(
        &self,
        request: Request<Streaming<protos::ftl::AcquireLeaseRequest>>,
    ) -> Result<Response<Self::AcquireLeaseStream>, Status> {
        todo!()
    }

    async fn send_fsm_event(
        &self,
        request: Request<protos::ftl::SendFsmEventRequest>,
    ) -> Result<Response<protos::ftl::SendFsmEventResponse>, Status> {
        todo!()
    }

    async fn publish_event(
        &self,
        request: Request<protos::ftl::PublishEventRequest>,
    ) -> Result<Response<protos::ftl::PublishEventResponse>, Status> {
        todo!()
    }

    async fn call(
        &self,
        request: Request<protos::ftl::CallRequest>,
    ) -> Result<Response<protos::ftl::CallResponse>, Status> {
        let request = request.into_inner();
        let verb_ref = request.verb.unwrap();
        let module = verb_ref.module;
        let name = verb_ref.name.to_snake_case();
        let request_body: Vec<u8> = request.body;
        let request_body = String::from_utf8(request_body).unwrap();

        let response =
            (self.config.call_immediate)(Context::default(), module, name, request_body).await;

        Ok(Response::new(protos::ftl::CallResponse {
            response: Some(protos::ftl::call_response::Response::Body(
                response.encode_to_vec(),
            )),
        }))
    }
}
