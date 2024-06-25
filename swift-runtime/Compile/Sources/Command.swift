import Foundation

struct CommandError: Error {
    let output:String
    let terminationStatus: Int32
}

func execute(command:String, directory:URL? = nil) throws -> String {
    let task = Process()
    let pipe = Pipe()
    let fileHandler = pipe.fileHandleForReading
    let errorPipe = Pipe()
    let errorFileHandler = errorPipe.fileHandleForReading
    
    if let directory = directory {
        task.currentDirectoryURL = directory
    }
        
    task.standardOutput = pipe
    task.standardError = errorPipe
    task.arguments = ["--login", // https://stackoverflow.com/a/45584225
                      "-c", command]
    
    // TODO: use bash
    // TODO: don't refer directly to executable path
    task.executableURL = URL(fileURLWithPath: "/bin/zsh")
    task.standardInput = nil
    
    let outputData = NSMutableData()
    fileHandler.readabilityHandler = { pipe in
        let data = pipe.availableData
        outputData.append(data)
    }
    
    let errorData = NSMutableData()
    errorFileHandler.readabilityHandler = { pipe in
        let data = pipe.availableData
        errorData.append(data)
    }
    
    try task.run()
    task.waitUntilExit()
    
    fileHandler.readabilityHandler = nil
    errorFileHandler.readabilityHandler = nil
    
    guard task.terminationStatus == 0 else {
        let errorOutput = String(data: errorData as Data, encoding: .utf8)!
        throw CommandError(output: errorOutput, terminationStatus: task.terminationStatus)
    }
    return String(data: outputData as Data, encoding:  .utf8)!
}
