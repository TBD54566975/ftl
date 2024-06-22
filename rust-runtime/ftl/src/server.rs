use std::future::Future;
use std::pin::Pin;

use tonic::{Request, Response, Status, Streaming};
use tonic::codegen::tokio_stream::Stream;
use tonic::transport::Server;

use ftl_protos as protos;
use ftl_protos::ftl::verb_service_server::VerbService;

use crate::Context;

#[derive(Debug)]
pub struct Config {
    pub call_immediate:
        fn(Context, String, String) -> Pin<Box<dyn Future<Output = ()> + Send + Sync>>,
}

pub async fn serve(config: Config) -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::1]:50051".parse()?;
    let service = FtlService { config };

    Server::builder()
        .add_service(protos::ftl::verb_service_server::VerbServiceServer::new(
            service,
        ))
        .serve(addr)
        .await?;

    Ok(())
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
        todo!()
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
        let name = verb_ref.name;

        (self.config.call_immediate)(Context::default(), module, name).await;

        Ok(Response::new(protos::ftl::CallResponse { response: None }))
    }
}
