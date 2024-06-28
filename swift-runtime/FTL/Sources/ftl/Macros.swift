import Foundation

/// Declares an FTL verb
@attached(peer, names: overloaded)
public macro FTLVerb() = #externalMacro(module: "FTLMacros", type: "VerbMacro")

/// Declares an FTL data type
@attached(extension, conformances: FTLType, names: arbitrary)
public macro FTLData() = #externalMacro(module: "FTLMacros", type: "DataMacro")

@freestanding(expression)
public macro callVerb<Req:FTLType, Resp:FTLType>(_ verb:Verb<Req,Resp>, request:Req) -> Resp = #externalMacro(module: "FTLMacros", type: "CallVerbMacro")

@freestanding(expression)
public macro callSource<Resp:FTLType>(_ source:Source<Resp>) -> Resp = #externalMacro(module: "FTLMacros", type: "CallSourceMacro")

@freestanding(expression)
public macro callSink<Req:FTLType>(_ sink:Sink<Req>, request:Req) = #externalMacro(module: "FTLMacros", type: "CallSinkMacro")

@freestanding(expression)
public macro callEmpty(_ emptyVerb:Empty) = #externalMacro(module: "FTLMacros", type: "CallEmptyMacro")
