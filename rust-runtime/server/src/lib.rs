// #[derive(Debug, Default)]
// pub struct MyGreeter {}
//
// #[tonic::async_trait]
// impl Greeter for MyGreeter {
//     async fn say_hello(
//         &self,
//         request: Request<HelloRequest>, // Accept request of type HelloRequest
//     ) -> Result<Response<HelloReply>, Status> {
//         // Return an instance of type HelloReply
//         println!("Got a request: {:?}", request);
//
//         let reply = HelloReply {
//             message: format!("Hello {}!", request.into_inner().name), // We must use .into_inner() as the fields of gRPC requests and responses are private
//         };
//
//         Ok(Response::new(reply)) // Send back our formatted greeting
//     }
// }

use std::pin::Pin;

use tonic::codegen::tokio_stream::Stream;
use tonic::transport::Server;
use tonic::{Request, Response, Status, Streaming};

use protos::ftl;
use protos::ftl::verb_service_server::VerbService;

pub async fn serve() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::1]:50051".parse()?;
    let service = FtlService::default();

    Server::builder()
        .add_service(protos::ftl::verb_service_server::VerbServiceServer::new(
            service,
        ))
        .serve(addr)
        .await?;

    Ok(())
}

#[derive(Debug, Default)]
pub struct FtlService {}

type ModuleContextResponseStream =
    Pin<Box<dyn Stream<Item = Result<ftl::ModuleContextResponse, Status>> + Send>>;

type AcquireLeaseStream =
    Pin<Box<dyn Stream<Item = Result<ftl::AcquireLeaseResponse, Status>> + Send>>;

#[tonic::async_trait]
impl VerbService for FtlService {
    async fn ping(
        &self,
        request: Request<ftl::PingRequest>,
    ) -> Result<Response<ftl::PingResponse>, Status> {
        todo!()
    }

    type GetModuleContextStream = ModuleContextResponseStream;

    async fn get_module_context(
        &self,
        request: Request<ftl::ModuleContextRequest>,
    ) -> Result<Response<Self::GetModuleContextStream>, Status> {
        todo!()
    }

    type AcquireLeaseStream = AcquireLeaseStream;

    async fn acquire_lease(
        &self,
        request: Request<Streaming<ftl::AcquireLeaseRequest>>,
    ) -> Result<Response<Self::AcquireLeaseStream>, Status> {
        todo!()
    }

    async fn send_fsm_event(
        &self,
        request: Request<ftl::SendFsmEventRequest>,
    ) -> Result<Response<ftl::SendFsmEventResponse>, Status> {
        todo!()
    }

    async fn publish_event(
        &self,
        request: Request<ftl::PublishEventRequest>,
    ) -> Result<Response<ftl::PublishEventResponse>, Status> {
        todo!()
    }

    async fn call(
        &self,
        request: Request<ftl::CallRequest>,
    ) -> Result<Response<ftl::CallResponse>, Status> {
        todo!()
    }
}
