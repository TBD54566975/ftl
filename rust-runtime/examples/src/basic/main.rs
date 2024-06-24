use tracing::info;

mod echo;

include!(concat!(env!("OUT_DIR"), "/call_immediate.rs"));

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    info!("Starting verb server");

    let bind = "localhost:1234".to_string();
    let bind_url = format!("http://{}", bind);

    let config = ftl::verb_server::Config {
        // call_immediate is a generated function that will call the appropriate verb.
        // See build.rs for how this is generated.
        bind,
        call_immediate,
    };

    // just as a test, we'll call the echo module's test_verb
    (config.call_immediate)(
        ftl::Context::default(),
        "echo".to_string(),
        "request_to_unit".to_string(),
        r#"{"name":"world","age":42}"#.to_string(),
    )
    .await;

    let controller_url = "http://localhost:8892".to_string();
    let config = ftl::runner::Config {
        verb_server_config: config,
        controller_url,
    };
    ftl::runner::run(config).await.unwrap();
}
