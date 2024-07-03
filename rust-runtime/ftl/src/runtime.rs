use tracing::info;

use ftl_protos::ftl::verb_service_client::VerbServiceClient;

use crate::verb_server;

#[derive(Debug)]
pub struct Config {
    pub verb_server_config: verb_server::Config,
    pub runner_url: String,
}

pub async fn run(config: Config) {
    info!("Starting");

    info!("Connecting to verb service at {}", config.runner_url);
    let verb_client = VerbServiceClient::connect(config.runner_url.clone())
        .await
        .unwrap();
    info!("Connected");

    verb_server::serve(config.verb_server_config, verb_client).await;
}
