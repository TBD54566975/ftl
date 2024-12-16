// @generated
// This file is @generated by prost-build.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(string, optional, tag="2")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="3")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(int32, tag="4")]
    pub log_level: i32,
    #[prost(map="string, string", tag="5")]
    pub attributes: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    #[prost(string, tag="6")]
    pub message: ::prost::alloc::string::String,
    #[prost(string, optional, tag="7")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(string, optional, tag="8")]
    pub stack: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CallEvent {
    #[prost(string, optional, tag="1")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(string, tag="2")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="11")]
    pub source_verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(message, optional, tag="12")]
    pub destination_verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(message, optional, tag="6")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(string, tag="7")]
    pub request: ::prost::alloc::string::String,
    #[prost(string, tag="8")]
    pub response: ::prost::alloc::string::String,
    #[prost(string, optional, tag="9")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(string, optional, tag="10")]
    pub stack: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeploymentCreatedEvent {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub language: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub module_name: ::prost::alloc::string::String,
    #[prost(int32, tag="4")]
    pub min_replicas: i32,
    #[prost(string, optional, tag="5")]
    pub replaced: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DeploymentUpdatedEvent {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(int32, tag="2")]
    pub min_replicas: i32,
    #[prost(int32, tag="3")]
    pub prev_min_replicas: i32,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IngressEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(string, optional, tag="2")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="3")]
    pub verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(string, tag="4")]
    pub method: ::prost::alloc::string::String,
    #[prost(string, tag="5")]
    pub path: ::prost::alloc::string::String,
    #[prost(int32, tag="7")]
    pub status_code: i32,
    #[prost(message, optional, tag="8")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="9")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(string, tag="10")]
    pub request: ::prost::alloc::string::String,
    #[prost(string, tag="11")]
    pub request_header: ::prost::alloc::string::String,
    #[prost(string, tag="12")]
    pub response: ::prost::alloc::string::String,
    #[prost(string, tag="13")]
    pub response_header: ::prost::alloc::string::String,
    #[prost(string, optional, tag="14")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CronScheduledEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(message, optional, tag="2")]
    pub verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(message, optional, tag="3")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="4")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(message, optional, tag="5")]
    pub scheduled_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(string, tag="6")]
    pub schedule: ::prost::alloc::string::String,
    #[prost(string, optional, tag="7")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AsyncExecuteEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(string, optional, tag="2")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="3")]
    pub verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(message, optional, tag="4")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="5")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(enumeration="AsyncExecuteEventType", tag="6")]
    pub async_event_type: i32,
    #[prost(string, optional, tag="7")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PubSubPublishEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(string, optional, tag="2")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="3")]
    pub verb_ref: ::core::option::Option<super::super::schema::v1::Ref>,
    #[prost(message, optional, tag="4")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="5")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(string, tag="6")]
    pub topic: ::prost::alloc::string::String,
    #[prost(string, tag="7")]
    pub request: ::prost::alloc::string::String,
    #[prost(string, optional, tag="8")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(int32, tag="9")]
    pub partition: i32,
    #[prost(int64, tag="10")]
    pub offset: i64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PubSubConsumeEvent {
    #[prost(string, tag="1")]
    pub deployment_key: ::prost::alloc::string::String,
    #[prost(string, optional, tag="2")]
    pub request_key: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(string, optional, tag="3")]
    pub dest_verb_module: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(string, optional, tag="4")]
    pub dest_verb_name: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(message, optional, tag="5")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag="6")]
    pub duration: ::core::option::Option<::prost_types::Duration>,
    #[prost(string, tag="7")]
    pub topic: ::prost::alloc::string::String,
    #[prost(string, optional, tag="8")]
    pub error: ::core::option::Option<::prost::alloc::string::String>,
    #[prost(int32, tag="9")]
    pub partition: i32,
    #[prost(int64, tag="10")]
    pub offset: i64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Event {
    #[prost(message, optional, tag="1")]
    pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
    /// Unique ID for event.
    #[prost(int64, tag="2")]
    pub id: i64,
    #[prost(oneof="event::Entry", tags="3, 4, 5, 6, 7, 8, 9, 10, 11")]
    pub entry: ::core::option::Option<event::Entry>,
}
/// Nested message and enum types in `Event`.
pub mod event {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Entry {
        #[prost(message, tag="3")]
        Log(super::LogEvent),
        #[prost(message, tag="4")]
        Call(super::CallEvent),
        #[prost(message, tag="5")]
        DeploymentCreated(super::DeploymentCreatedEvent),
        #[prost(message, tag="6")]
        DeploymentUpdated(super::DeploymentUpdatedEvent),
        #[prost(message, tag="7")]
        Ingress(super::IngressEvent),
        #[prost(message, tag="8")]
        CronScheduled(super::CronScheduledEvent),
        #[prost(message, tag="9")]
        AsyncExecute(super::AsyncExecuteEvent),
        #[prost(message, tag="10")]
        PubsubPublish(super::PubSubPublishEvent),
        #[prost(message, tag="11")]
        PubsubConsume(super::PubSubConsumeEvent),
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum EventType {
    Unspecified = 0,
    Log = 1,
    Call = 2,
    DeploymentCreated = 3,
    DeploymentUpdated = 4,
    Ingress = 5,
    CronScheduled = 6,
    AsyncExecute = 7,
    PubsubPublish = 8,
    PubsubConsume = 9,
}
impl EventType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Self::Unspecified => "EVENT_TYPE_UNSPECIFIED",
            Self::Log => "EVENT_TYPE_LOG",
            Self::Call => "EVENT_TYPE_CALL",
            Self::DeploymentCreated => "EVENT_TYPE_DEPLOYMENT_CREATED",
            Self::DeploymentUpdated => "EVENT_TYPE_DEPLOYMENT_UPDATED",
            Self::Ingress => "EVENT_TYPE_INGRESS",
            Self::CronScheduled => "EVENT_TYPE_CRON_SCHEDULED",
            Self::AsyncExecute => "EVENT_TYPE_ASYNC_EXECUTE",
            Self::PubsubPublish => "EVENT_TYPE_PUBSUB_PUBLISH",
            Self::PubsubConsume => "EVENT_TYPE_PUBSUB_CONSUME",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "EVENT_TYPE_UNSPECIFIED" => Some(Self::Unspecified),
            "EVENT_TYPE_LOG" => Some(Self::Log),
            "EVENT_TYPE_CALL" => Some(Self::Call),
            "EVENT_TYPE_DEPLOYMENT_CREATED" => Some(Self::DeploymentCreated),
            "EVENT_TYPE_DEPLOYMENT_UPDATED" => Some(Self::DeploymentUpdated),
            "EVENT_TYPE_INGRESS" => Some(Self::Ingress),
            "EVENT_TYPE_CRON_SCHEDULED" => Some(Self::CronScheduled),
            "EVENT_TYPE_ASYNC_EXECUTE" => Some(Self::AsyncExecute),
            "EVENT_TYPE_PUBSUB_PUBLISH" => Some(Self::PubsubPublish),
            "EVENT_TYPE_PUBSUB_CONSUME" => Some(Self::PubsubConsume),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum AsyncExecuteEventType {
    Unspecified = 0,
    Cron = 1,
    Pubsub = 2,
}
impl AsyncExecuteEventType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Self::Unspecified => "ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED",
            Self::Cron => "ASYNC_EXECUTE_EVENT_TYPE_CRON",
            Self::Pubsub => "ASYNC_EXECUTE_EVENT_TYPE_PUBSUB",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED" => Some(Self::Unspecified),
            "ASYNC_EXECUTE_EVENT_TYPE_CRON" => Some(Self::Cron),
            "ASYNC_EXECUTE_EVENT_TYPE_PUBSUB" => Some(Self::Pubsub),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum LogLevel {
    Unspecified = 0,
    Trace = 1,
    Debug = 5,
    Info = 9,
    Warn = 13,
    Error = 17,
}
impl LogLevel {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            Self::Unspecified => "LOG_LEVEL_UNSPECIFIED",
            Self::Trace => "LOG_LEVEL_TRACE",
            Self::Debug => "LOG_LEVEL_DEBUG",
            Self::Info => "LOG_LEVEL_INFO",
            Self::Warn => "LOG_LEVEL_WARN",
            Self::Error => "LOG_LEVEL_ERROR",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "LOG_LEVEL_UNSPECIFIED" => Some(Self::Unspecified),
            "LOG_LEVEL_TRACE" => Some(Self::Trace),
            "LOG_LEVEL_DEBUG" => Some(Self::Debug),
            "LOG_LEVEL_INFO" => Some(Self::Info),
            "LOG_LEVEL_WARN" => Some(Self::Warn),
            "LOG_LEVEL_ERROR" => Some(Self::Error),
            _ => None,
        }
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTimelineRequest {
    #[prost(message, repeated, tag="1")]
    pub filters: ::prost::alloc::vec::Vec<get_timeline_request::Filter>,
    #[prost(int32, tag="2")]
    pub limit: i32,
    /// Ordering is done by id which matches publication order.
    /// This roughly corresponds to the time of the event, but not strictly.
    #[prost(enumeration="get_timeline_request::Order", tag="3")]
    pub order: i32,
}
/// Nested message and enum types in `GetTimelineRequest`.
pub mod get_timeline_request {
    /// Filters events by log level.
    #[derive(Clone, Copy, PartialEq, ::prost::Message)]
    pub struct LogLevelFilter {
        #[prost(enumeration="super::LogLevel", tag="1")]
        pub log_level: i32,
    }
    /// Filters events by deployment key.
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct DeploymentFilter {
        #[prost(string, repeated, tag="1")]
        pub deployments: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    }
    /// Filters events by request key.
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct RequestFilter {
        #[prost(string, repeated, tag="1")]
        pub requests: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    }
    /// Filters events by event type.
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct EventTypeFilter {
        #[prost(enumeration="super::EventType", repeated, tag="1")]
        pub event_types: ::prost::alloc::vec::Vec<i32>,
    }
    /// Filters events by time.
    ///
    /// Either end of the time range can be omitted to indicate no bound.
    #[derive(Clone, Copy, PartialEq, ::prost::Message)]
    pub struct TimeFilter {
        #[prost(message, optional, tag="1")]
        pub older_than: ::core::option::Option<::prost_types::Timestamp>,
        #[prost(message, optional, tag="2")]
        pub newer_than: ::core::option::Option<::prost_types::Timestamp>,
    }
    /// Filters events by ID.
    ///
    /// Either end of the ID range can be omitted to indicate no bound.
    #[derive(Clone, Copy, PartialEq, ::prost::Message)]
    pub struct IdFilter {
        #[prost(int64, optional, tag="1")]
        pub lower_than: ::core::option::Option<i64>,
        #[prost(int64, optional, tag="2")]
        pub higher_than: ::core::option::Option<i64>,
    }
    /// Filters events by call.
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct CallFilter {
        #[prost(string, tag="1")]
        pub dest_module: ::prost::alloc::string::String,
        #[prost(string, optional, tag="2")]
        pub dest_verb: ::core::option::Option<::prost::alloc::string::String>,
        #[prost(string, optional, tag="3")]
        pub source_module: ::core::option::Option<::prost::alloc::string::String>,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ModuleFilter {
        #[prost(string, tag="1")]
        pub module: ::prost::alloc::string::String,
        #[prost(string, optional, tag="2")]
        pub verb: ::core::option::Option<::prost::alloc::string::String>,
    }
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct Filter {
        /// These map 1:1 with filters in backend/timeline/filters.go
        #[prost(oneof="filter::Filter", tags="1, 2, 3, 4, 5, 6, 7, 8")]
        pub filter: ::core::option::Option<filter::Filter>,
    }
    /// Nested message and enum types in `Filter`.
    pub mod filter {
        /// These map 1:1 with filters in backend/timeline/filters.go
        #[derive(Clone, PartialEq, ::prost::Oneof)]
        pub enum Filter {
            #[prost(message, tag="1")]
            LogLevel(super::LogLevelFilter),
            #[prost(message, tag="2")]
            Deployments(super::DeploymentFilter),
            #[prost(message, tag="3")]
            Requests(super::RequestFilter),
            #[prost(message, tag="4")]
            EventTypes(super::EventTypeFilter),
            #[prost(message, tag="5")]
            Time(super::TimeFilter),
            #[prost(message, tag="6")]
            Id(super::IdFilter),
            #[prost(message, tag="7")]
            Call(super::CallFilter),
            #[prost(message, tag="8")]
            Module(super::ModuleFilter),
        }
    }
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Order {
        Unspecified = 0,
        Asc = 1,
        Desc = 2,
    }
    impl Order {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Self::Unspecified => "ORDER_UNSPECIFIED",
                Self::Asc => "ORDER_ASC",
                Self::Desc => "ORDER_DESC",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "ORDER_UNSPECIFIED" => Some(Self::Unspecified),
                "ORDER_ASC" => Some(Self::Asc),
                "ORDER_DESC" => Some(Self::Desc),
                _ => None,
            }
        }
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTimelineResponse {
    #[prost(message, repeated, tag="1")]
    pub events: ::prost::alloc::vec::Vec<Event>,
    /// For pagination, this cursor is where we should start our next query
    #[prost(int64, optional, tag="2")]
    pub cursor: ::core::option::Option<i64>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StreamTimelineRequest {
    #[prost(message, optional, tag="1")]
    pub update_interval: ::core::option::Option<::prost_types::Duration>,
    #[prost(message, optional, tag="2")]
    pub query: ::core::option::Option<GetTimelineRequest>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StreamTimelineResponse {
    #[prost(message, repeated, tag="1")]
    pub events: ::prost::alloc::vec::Vec<Event>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateEventsRequest {
    #[prost(message, repeated, tag="1")]
    pub entries: ::prost::alloc::vec::Vec<create_events_request::EventEntry>,
}
/// Nested message and enum types in `CreateEventsRequest`.
pub mod create_events_request {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct EventEntry {
        #[prost(message, optional, tag="1")]
        pub timestamp: ::core::option::Option<::prost_types::Timestamp>,
        #[prost(oneof="event_entry::Entry", tags="2, 3, 4, 5, 6, 7, 8, 9, 10")]
        pub entry: ::core::option::Option<event_entry::Entry>,
    }
    /// Nested message and enum types in `EventEntry`.
    pub mod event_entry {
        #[derive(Clone, PartialEq, ::prost::Oneof)]
        pub enum Entry {
            #[prost(message, tag="2")]
            Log(super::super::LogEvent),
            #[prost(message, tag="3")]
            Call(super::super::CallEvent),
            #[prost(message, tag="4")]
            DeploymentCreated(super::super::DeploymentCreatedEvent),
            #[prost(message, tag="5")]
            DeploymentUpdated(super::super::DeploymentUpdatedEvent),
            #[prost(message, tag="6")]
            Ingress(super::super::IngressEvent),
            #[prost(message, tag="7")]
            CronScheduled(super::super::CronScheduledEvent),
            #[prost(message, tag="8")]
            AsyncExecute(super::super::AsyncExecuteEvent),
            #[prost(message, tag="9")]
            PubsubPublish(super::super::PubSubPublishEvent),
            #[prost(message, tag="10")]
            PubsubConsume(super::super::PubSubConsumeEvent),
        }
    }
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct CreateEventsResponse {
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct DeleteOldEventsRequest {
    #[prost(enumeration="EventType", tag="1")]
    pub event_type: i32,
    #[prost(int64, tag="2")]
    pub age_seconds: i64,
}
#[derive(Clone, Copy, PartialEq, ::prost::Message)]
pub struct DeleteOldEventsResponse {
    #[prost(int64, tag="1")]
    pub deleted_count: i64,
}
// @@protoc_insertion_point(module)