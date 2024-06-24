import Foundation
import ArgumentParser
import Schema

enum CompileError: Error {
    case serializeSchema(Error)
}

@main
struct Compile: ParsableCommand {
    @Argument(help: "module name")
    var name: String
    
    @Argument(help: "module root directory")
    var rootPath: String
    
    @Argument(help: "command to build the module")
    var buildCmd: String
    
    @Argument(help: "directory path to deploy into")
    var deployPath: String
    
    @Argument(help: "schema filename in the deploy path")
    var schemaFilename: String
    
    mutating func run() throws {
        let module = try Extractor(rootPath:self.rootPath).extract()
        do {
            let data = try module.serializedBytes()
            let schemaUrl = URL(filePath: self.deployPath).appending(path: schemaFilename)
            try data.write(to: schemaUrl)
        }
        catch {
            throw CompileError.serializeSchema(error)
        }
        
        
    }
}
