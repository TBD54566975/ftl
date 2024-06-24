// The Swift Programming Language
// https://docs.swift.org/swift-book

/// Declares an FTL verb
@attached(peer, names: overloaded)
public macro FTLVerb() = #externalMacro(module: "FTLMacros", type: "VerbMacro")

/// Declares an FTL data type
@attached(member, names: overloaded)
public macro FTLData() = #externalMacro(module: "FTLMacros", type: "DataMacro")
