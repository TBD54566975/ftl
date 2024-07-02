import Foundation

import ArgumentParser
import GRPC
import NIOCore
import NIOPosix
import Stime
import FTL

struct MainError: Error {
   let message: String
}

@main
struct Main: ParsableCommand {
   mutating func run() throws {
      let group = MultiThreadedEventLoopGroup(numberOfThreads: System.coreCount)
      defer {
         try! group.syncShutdownGracefully()
      }
      
      guard let bindURLString = ProcessInfo().environment["FTL_BIND"],
            let bindURL = URL(string:bindURLString),
            let host = bindURL.host,
            let port = bindURL.port else {
         throw MainError(message: "could not parse host and port to bind to: \(ProcessInfo().environment["FTL_BIND"] ?? "No envar value for FTL_BIND")")
      }
      
      let client = try Client(group: group, host:host, port:port)
      let context = FTL.Context(client:client)
      let provider = VerbServiceProvider(context, handlers: [
{{- range .Verbs}}
handlerFor(name: "{{.Name}}", {{.Package}}.{{.Name}}),
{{- end}}
      ])
      
      // TODO: reconsider plaintext transport
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
