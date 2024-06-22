use tracing::info;

use ftl_protos::ftl::verb_service_client::VerbServiceClient;
use ftl_protos::schema::Ref;

pub async fn call_verb(module: String, name: String) {
    info!("Calling verb {} in module {}", name, module);

    let mut client = VerbServiceClient::connect("http://[::1]:50051")
        .await
        .unwrap();
    let request = tonic::Request::new(ftl_protos::ftl::CallRequest {
        metadata: None,
        verb: Some(Ref {
            pos: None,
            name,
            module,
            type_parameters: vec![],
        }),
        body: vec![],
    });

    let response = client.call(request).await.unwrap();
    info!("Response: {:?}", response);
}
