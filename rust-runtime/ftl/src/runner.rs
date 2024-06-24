//! Communicates with the controller and manages the verb server.

use tonic::{Request, Response, Status};
use tonic::codegen::tokio_stream;
use tonic::transport::Server;
use tracing::{info, trace};

use ftl_protos::ftl::{
    call_response, CallRequest, DeployRequest, DeployResponse, GetSchemaRequest, PingRequest,
    PingResponse, RegisterRunnerRequest, ReserveRequest, ReserveResponse, TerminateRequest,
};
use ftl_protos::ftl::controller_service_client::ControllerServiceClient;
use ftl_protos::ftl::runner_service_server::RunnerServiceServer;
use ftl_protos::ftl::verb_service_client::VerbServiceClient;
use ftl_protos::schema::{decl, Decl, Ref, Verb};

use crate::verb_server;

pub struct Config {
    pub verb_server_config: verb_server::Config,
    pub controller_url: String,
}

pub async fn run(config: Config) -> Result<(), Box<dyn std::error::Error>> {
    info!("Starting runner");

    info!(
        "Connecting to controller ftl service at {}",
        config.controller_url
    );
    let mut controller_client =
        ControllerServiceClient::connect(config.controller_url.clone()).await?;
    info!("Connected to controller");

    info!(
        "Connecting to controller verb service at {}",
        config.controller_url
    );
    let mut verb_client = VerbServiceClient::connect(config.controller_url.clone()).await?;
    info!("Connected to controller");

    tokio::task::spawn(async move {
        let runner_service = RunnerService {};
        Server::builder()
            .add_service(RunnerServiceServer::new(runner_service))
            .serve("127.0.0.1:1234".parse().unwrap())
            .await
            .unwrap();
    });

    let response = controller_client
        .get_schema(GetSchemaRequest {})
        .await
        .unwrap();
    let message = response.get_ref().clone();
    for module in message.schema.unwrap().modules {
        info!("Module: {}", module.name);
        for d in module.decls {
            match d.value.unwrap() {
                decl::Value::Verb(verb) => {
                    info!(" - Verb: {}", verb.name);
                    let request_type = verb.request.unwrap().value.unwrap();
                    let response_type = verb.response.unwrap().value.unwrap();
                    info!("   - Request: {:?}", request_type);
                    info!("   - Response: {:?}", response_type);
                }
                decl::Value::Data(data) => {
                    info!(" - Data: {:?}", data);
                }
                _ => {}
            }
        }
    }

    // just attempting to call a verb on the controller manually
    let response = verb_client
        .call(CallRequest {
            metadata: None,
            verb: Some(Ref {
                pos: None,
                name: "time".to_string(),
                module: "time".to_string(),
                type_parameters: vec![],
            }),
            body: r#"{}"#.as_bytes().to_vec(),
        })
        .await
        .unwrap();

    info!("..................");
    info!("Response: {:?}", response);
    let message = response.get_ref().clone();
    let response = message.response.unwrap();
    match response {
        call_response::Response::Body(body) => {
            info!("Body: {}", String::from_utf8(body).unwrap());
        }
        call_response::Response::Error(e) => {
            info!("Error: {:?}", e);
        }
    }

    // tokio::task::spawn(verb_server::serve(config.verb_server_config));

    registration_loop(
        controller_client.clone(),
        config.verb_server_config.bind_url(),
    )
    .await;

    Ok(())
}

async fn registration_loop(
    mut controller_client: ControllerServiceClient<tonic::transport::channel::Channel>,
    string: String,
) {
    loop {
        info!("Registering runner");
        let response = controller_client
            .register_runner(tokio_stream::iter([RegisterRunnerRequest {
                key: "rnr-1-4wv6bcqc3pfka8bk".to_string(),
                endpoint: string.clone(),
                deployment: None,
                state: 0,
                labels: None,
                error: None,
            }]))
            .await
            .unwrap();

        info!("Registered runner: {:?}", response);

        tokio::time::sleep(std::time::Duration::from_secs(5)).await;
    }
}

#[derive(Debug)]
pub struct RunnerService {}

#[tonic::async_trait]
impl ftl_protos::ftl::runner_service_server::RunnerService for RunnerService {
    async fn ping(&self, request: Request<PingRequest>) -> Result<Response<PingResponse>, Status> {
        trace!("Ping request: {:?}", request);
        Ok(Response::new(PingResponse { not_ready: None }))
    }

    async fn reserve(
        &self,
        request: Request<ReserveRequest>,
    ) -> Result<Response<ReserveResponse>, Status> {
        todo!("reserve")
    }

    async fn deploy(
        &self,
        request: Request<DeployRequest>,
    ) -> Result<Response<DeployResponse>, Status> {
        todo!("deploy")
    }

    async fn terminate(
        &self,
        request: Request<TerminateRequest>,
    ) -> Result<Response<RegisterRunnerRequest>, Status> {
        todo!("terminate")
    }
}
