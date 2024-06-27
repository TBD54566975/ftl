// The Swift Programming Language
// https://docs.swift.org/swift-book

/// Declares an FTL verb
@attached(peer, names: overloaded)
public macro FTLVerb() = #externalMacro(module: "FTLMacros", type: "VerbMacro")

/// Declares an FTL data type
@attached(extension, conformances: FTLType, names: arbitrary)
public macro FTLData() = #externalMacro(module: "FTLMacros", type: "DataMacro")

public protocol FTLType {
   static func ftlDecode(_ json:Any?) throws -> Self
   func ftlEncode() -> Any?
}
