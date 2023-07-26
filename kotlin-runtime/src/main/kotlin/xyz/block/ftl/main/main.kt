package xyz.block.ftl.main

import io.grpc.ServerBuilder
import io.grpc.netty.NettyServerBuilder
import xyz.block.ftl.registry.Registry
import xyz.block.ftl.server.Server
import java.net.InetSocketAddress
import java.net.URL

val defaultBindAddress = "http://127.0.0.1:8443"

fun main() {
  val bind = URL(System.getenv("FTL_BIND") ?: defaultBindAddress)
  val addr = InetSocketAddress(bind.host, bind.port)
  val registry = Registry()
  registry.registerAll()
  val server = Server(registry)
  val grpcServer = NettyServerBuilder.forAddress(addr)
    .addService(server)
    .build()
  grpcServer.start()
  grpcServer.awaitTermination()
}