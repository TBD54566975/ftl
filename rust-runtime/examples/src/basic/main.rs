mod echo;

include!(concat!(env!("OUT_DIR"), "/lookup.rs"));

#[tokio::main]
async fn main() {
    let config = ftl::server::Config {
        // call_immediate comes from the generated lookup.rs
        call_immediate,
    };

    // just as a test, we'll call the echo module's test_verb
    (config.call_immediate)(
        ftl::Context::default(),
        "echo".to_string(),
        "test_verb".to_string(),
    )
    .await;

    ftl::serve(config).await.unwrap();
}
