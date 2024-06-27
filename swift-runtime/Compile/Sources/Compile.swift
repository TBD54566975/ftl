import Foundation
import ArgumentParser
import Schema

enum CompileError: Error {
    case serializeSchema(Error)
}

@main
struct Compile: ParsableCommand {
    @Option(help: "module name")
    var name: String
    
    @Option(help: "module root directory")
    var rootPath: String
    
    @Option(help: "directory path to deploy into")
    var deployPath: String
    
    @Option(help: "schema filename in the deploy path")
    var schemaFilename: String
    
    mutating func run() throws {
        let rootURL = URL(fileURLWithPath:self.rootPath)
        let deployURL = URL(fileURLWithPath:self.deployPath)
        
        let module = try Extractor(rootURL: rootURL, deployURL:deployURL, name: self.name).extract()
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
