import Foundation
import SwiftSyntax
import SwiftParser

/// Extractor takes a module directory and extracts an FTL schema
public class Extractor {
    let rootPath: String
    let name: String
    
    public init(rootPath: String, name:String) {
        self.rootPath = rootPath
        self.name = name
    }
    
    public func extract() throws -> Module {
        let rootURL = URL(filePath: self.rootPath)
        var module = Module(name: name)
        module.comments.append("Fake it till we make it")
        SwiftParser.Parser(
        return module
    }
}
