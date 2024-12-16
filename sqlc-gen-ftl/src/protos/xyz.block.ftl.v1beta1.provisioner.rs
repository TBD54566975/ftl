// @generated
// This file is @generated by prost-build.
/// Resource is an abstract resource extracted from FTL Schema.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Resource {
    /// id unique within the module
    #[prost(string, tag="1")]
    pub resource_id: ::prost::alloc::string::String,
    #[prost(oneof="resource::Resource", tags="102, 103, 104")]
    pub resource: ::core::option::Option<resource::Resource>,
}
/// Nested message and enum types in `Resource`.
pub mod resource {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Resource {
        #[prost(message, tag="102")]
        Postgres(super::PostgresResource),
        #[prost(message, tag="103")]
        Mysql(super::MysqlResource),
        #[prost(message, tag="104")]
        Module(super::ModuleResource),
    }
}
// Resource types
//
// any output created by the provisioner is stored in a field called "output"

#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PostgresResource {
    #[prost(message, optional, tag="1")]
    pub output: ::core::option::Option<postgres_resource::PostgresResourceOutput>,
}
/// Nested message and enum types in `PostgresResource`.
pub mod postgres_resource {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct PostgresResourceOutput {
        #[prost(string, tag="1")]
        pub read_dsn: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub write_dsn: ::prost::alloc::string::String,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MysqlResource {
    #[prost(message, optional, tag="1")]
    pub output: ::core::option::Option<mysql_resource::MysqlResourceOutput>,
}
/// Nested message and enum types in `MysqlResource`.
pub mod mysql_resource {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct MysqlResourceOutput {
        #[prost(string, tag="1")]
        pub read_dsn: ::prost::alloc::string::String,
        #[prost(string, tag="2")]
        pub write_dsn: ::prost::alloc::string::String,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ModuleResource {
    #[prost(message, optional, tag="1")]
    pub output: ::core::option::Option<module_resource::ModuleResourceOutput>,
    #[prost(message, optional, tag="2")]
    pub schema: ::core::option::Option<super::super::v1::schema::Module>,
    #[prost(message, repeated, tag="3")]
    pub artefacts: ::prost::alloc::vec::Vec<super::super::v1::DeploymentArtefact>,
    /// Runner labels required to run this deployment.
    #[prost(message, optional, tag="4")]
    pub labels: ::core::option::Option<::prost_types::Struct>,
}
/// Nested message and enum types in `ModuleResource`.
pub mod module_resource {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ModuleResourceOutput {
        #[prost(string, tag="1")]
        pub deployment_key: ::prost::alloc::string::String,
    }
}
/// ResourceContext is the context used to create a new resource
/// This includes the direct dependencies of the new resource, that can impact
/// the resource creation.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ResourceContext {
    #[prost(message, optional, tag="1")]
    pub resource: ::core::option::Option<Resource>,
    #[prost(message, repeated, tag="2")]
    pub dependencies: ::prost::alloc::vec::Vec<Resource>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProvisionRequest {
    #[prost(string, tag="1")]
    pub ftl_cluster_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub module: ::prost::alloc::string::String,
    /// The resource FTL thinks exists currently
    #[prost(message, repeated, tag="3")]
    pub existing_resources: ::prost::alloc::vec::Vec<Resource>,
    /// The resource FTL would like to exist after this provisioning run.
    /// This includes all new, existing, and changes resources in this change.
    #[prost(message, repeated, tag="4")]
    pub desired_resources: ::prost::alloc::vec::Vec<ResourceContext>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ProvisionResponse {
    #[prost(string, tag="1")]
    pub provisioning_token: ::prost::alloc::string::String,
    #[prost(enumeration="provision_response::ProvisionResponseStatus", tag="2")]
    pub status: i32,
}
/// Nested message and enum types in `ProvisionResponse`.
pub mod provision_response {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum ProvisionResponseStatus {
        Unknown = 0,
        Submitted = 1,
    }
    impl ProvisionResponseStatus {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Self::Unknown => "UNKNOWN",
                Self::Submitted => "SUBMITTED",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNKNOWN" => Some(Self::Unknown),
                "SUBMITTED" => Some(Self::Submitted),
                _ => None,
            }
        }
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StatusRequest {
    #[prost(string, tag="1")]
    pub provisioning_token: ::prost::alloc::string::String,
    /// The set of desired_resources used to initiate this provisioning request
    /// We need this as input here, so we can populate any resource fields in them
    /// when the provisioning finishes
    #[prost(message, repeated, tag="2")]
    pub desired_resources: ::prost::alloc::vec::Vec<Resource>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StatusResponse {
    #[prost(oneof="status_response::Status", tags="1, 2")]
    pub status: ::core::option::Option<status_response::Status>,
}
/// Nested message and enum types in `StatusResponse`.
pub mod status_response {
    #[derive(Clone, Copy, PartialEq, ::prost::Message)]
    pub struct ProvisioningRunning {
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ProvisioningFailed {
        #[prost(string, tag="1")]
        pub error_message: ::prost::alloc::string::String,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ProvisioningSuccess {
        /// Some fields in the resources might have been populated
        /// during the provisioning. The new state is returned here
        #[prost(message, repeated, tag="1")]
        pub updated_resources: ::prost::alloc::vec::Vec<super::Resource>,
    }
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Status {
        #[prost(message, tag="1")]
        Running(ProvisioningRunning),
        #[prost(message, tag="2")]
        Success(ProvisioningSuccess),
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PlanRequest {
    #[prost(message, optional, tag="1")]
    pub provisioning: ::core::option::Option<ProvisionRequest>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PlanResponse {
    /// a detailed, implementation specific, plan of changes this deployment would do
    #[prost(string, tag="1")]
    pub plan: ::prost::alloc::string::String,
}
// @@protoc_insertion_point(module)