//! Communicates with the controller and manages the verb server.

use std::path::Path;

use clap::Parser;
use tonic::{Request, Response, Status};
use tonic::codegen::tokio_stream;
use tonic::transport::Server;
use tracing::{debug, info, trace};
use tracing_subscriber::{EnvFilter, fmt};

use ftl_protos::ftl::{
    call_response, CallRequest, DeployRequest, DeployResponse, GetSchemaRequest, PingRequest,
    PingResponse, ReserveRequest, ReserveResponse, TerminateRequest,
};
use ftl_protos::ftl::controller_service_client::ControllerServiceClient;
use ftl_protos::ftl::verb_service_client::VerbServiceClient;
use ftl_protos::schema::{decl, Decl, Ref, Verb};

use crate::{parser, verb_server};
use crate::verb_server::CallImmediateFn;

#[derive(Debug)]
pub struct Config {
    pub verb_server_config: verb_server::Config,
    pub runner_url: String,
}

pub async fn run(config: Config) -> () {
    info!("Starting");

    info!("Connecting to verb service at {}", config.runner_url);
    let mut verb_client = VerbServiceClient::connect(config.runner_url.clone())
        .await
        .unwrap();
    info!("Connected");

    verb_server::serve(config.verb_server_config, verb_client).await;
}
