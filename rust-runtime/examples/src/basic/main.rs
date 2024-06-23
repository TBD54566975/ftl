use tracing::info;

mod echo;

include!(concat!(env!("OUT_DIR"), "/call_immediate.rs"));

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    info!("Starting server");

    let config = ftl::server::Config {
        // call_immediate is a generated function that will call the appropriate verb.
        // See build.rs for how this is generated.
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

    ftl::serve(config).await.unwrap();
}
