import Foundation
import ArgumentParser
import GRPC
import NIOCore
import NIOPosix
import Stime

struct MainError: Error {
   let message: String
}

@main
struct Main: ParsableCommand {
   //    @Option(help: "socket to bind to")
   //    var bind: String
   
   mutating func run() throws {
      let group = MultiThreadedEventLoopGroup(numberOfThreads: System.coreCount)
      defer {
         try! group.syncShutdownGracefully()
      }
      
      let provider = VerbServiceProvider(handlers: [
         try handlerFor(name: "time", Stime.time)
      ])
      
      guard let bindURLString = ProcessInfo().environment["FTL_BIND"],
            let bindURL = URL(string:bindURLString),
            let host = bindURL.host,
            let port = bindURL.port else {
         throw MainError(message: "could not parse host and port to bind to: \(ProcessInfo().environment["FTL_BIND"])")
      }
      let server = Server.insecure(group: group).withServiceProviders([provider]).bind(host: host, port: port)
      server.map {
         $0.channel.localAddress
      }.whenSuccess { address in
         print("swift ftl server started on port \(address!.port!)")
      }
      
      _ = try server.flatMap {
         $0.onClose
      }.wait()
   }
}

func fakeVerb(_ req:String) throws -> (Int) {
   return 0
}

