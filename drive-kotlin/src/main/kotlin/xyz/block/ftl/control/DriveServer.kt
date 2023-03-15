package xyz.block.ftl.control

import ch.qos.logback.classic.Logger
import io.grpc.ManagedChannelBuilder
import io.grpc.netty.NettyServerBuilder
import xyz.block.ftl.drive.verb.VerbDeck
import xyz.block.ftl.v1.VerbServiceGrpcKt
import java.net.InetSocketAddress

/**
 * Start FTL drive-kotlin.
 */
fun startDrive(logger: Logger, deck: VerbDeck) {
  // Create client
  val ftlEndpoint = parseSocket(System.getenv("FTL_ENDPOINT"))
  var routerClient: VerbServiceGrpcKt.VerbServiceCoroutineStub? = null
  if (ftlEndpoint != null && ftlEndpoint is InetSocketAddress) {
    val channel = ManagedChannelBuilder
      .forAddress(ftlEndpoint.hostName, ftlEndpoint.port)
      .usePlaintext()
      .build()
    routerClient = VerbServiceGrpcKt.VerbServiceCoroutineStub(channel)
  }

  // Start the gRPC servers.
  val pluginEndpoint = parseSocket(System.getenv("FTL_PLUGIN_ENDPOINT") ?: "127.0.0.1:8081")!!
  logger.info("Starting FTL.kotlin-drive gRPC services on $pluginEndpoint")

  val server = NettyServerBuilder
    .forAddress(pluginEndpoint)
    .addService(VerbServer(deck))
    .addService(DevelServer())
    .build()
  // TODO: Terminate the process if this fails to startup.
  server.start()
}