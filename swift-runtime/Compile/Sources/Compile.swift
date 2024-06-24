import ArgumentParser

@main
struct compile: ParsableCommand {
    @Argument(help: "Existing schema")
       var schemaString: String
    
    @Argument(help: "modulePath")
       var modulePath: String
    
    mutating func run() throws {
        print("Hello, world!\n\(modulePath)\n\(schemaString)")
    }
}
