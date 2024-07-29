package xyz.block.ftl.main

import io.grpc.ServerInterceptors
import io.grpc.netty.NettyServerBuilder
import xyz.block.ftl.client.GrpcVerbServiceClient
import xyz.block.ftl.client.makeGrpcClient
import xyz.block.ftl.registry.Registry
import xyz.block.ftl.server.Server
import xyz.block.ftl.server.ServerInterceptor
import xyz.block.ftl.v1.VerbServiceWireGrpc.VerbServiceBlockingStub
import java.net.InetSocketAddress
import java.net.URL

val defaultBindAddress = "http://127.0.0.1:8896"
val defaultFtlEndpoint = "http://127.0.0.1:8892"

fun main() {
  val bind = URL(System.getenv("FTL_BIND") ?: defaultBindAddress)
  val addr = InetSocketAddress(bind.host, bind.port)
  val registry = Registry()
  registry.registerAll()
  for (verb in registry.refs) {
    println("Registered verb: ${verb.module}.${verb.name}")
  }
  val ftlEndpoint = System.getenv("FTL_ENDPOINT") ?: defaultFtlEndpoint
  val grpcClient = VerbServiceBlockingStub(makeGrpcClient(ftlEndpoint))
  val verbRoutingClient = GrpcVerbServiceClient(grpcClient)
  val server = Server(registry, verbRoutingClient)
  val grpcServer = NettyServerBuilder.forAddress(addr)
    .addService(ServerInterceptors.intercept(server, ServerInterceptor()))
    .build()
  grpcServer.start()
  grpcServer.awaitTermination()
}
