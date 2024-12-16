// @generated
// This file is @generated by prost-build.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PublishEventRequest {
    #[prost(message, optional, tag="1")]
    pub topic: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(bytes="bytes", tag="2")]
    pub body: ::prost::bytes::Bytes,
    #[prost(string, tag="3")]
    pub key: ::prost::alloc::string::String,
    /// Only verb name is included because this verb will be in the same module as topic
    #[prost(string, tag="4")]
    pub caller: ::prost::alloc::string::String,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct PublishEventResponse {
}
// @@protoc_insertion_point(module)