package xyz.block.ftl.drive

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.server.ServerConnector
import org.eclipse.jetty.servlet.ServletHandler
import xyz.block.ftl.control.DevelServer
import xyz.block.ftl.control.VerbServer
import xyz.block.ftl.control.parseSocket
import xyz.block.ftl.control.startControlChannelServer
import xyz.block.ftl.drive.transport.DriveServlet
import xyz.block.ftl.drive.verb.VerbDeck

val messages = listOf(
  "Warming up dilithium chamber...",
  "Initializing warp core...",
  "Sparking matter/anti-matter reactor...",
  "Engaging Proto-Star Drive...",
  "Connecting to the Mycelial Network..."
)

fun main(args: Array<String>) {
  Logging.init()
  val logger = Logging.logger("FTL Drive")
  logger.info(messages[(Math.random() * 10 % messages.size).toInt()])

  val server = Server()
  server.connectors = arrayOf(ServerConnector(server).apply {
    port = 8080
  })
  server.handler = ServletHandler().apply {
    addServletWithMapping(DriveServlet::class.java, "/")
  }
  VerbDeck.init("com.squareup.ftldemo")

  // Start control channel if requested.
  val pluginSocket = System.getenv("FTL_PLUGIN_SOCKET")
  if (pluginSocket != null) {
    val controlChannelSocket = parseSocket(pluginSocket)
    logger.info("Listening to Drive control channel on $controlChannelSocket")

    startControlChannelServer(
      controlChannelSocket,
      VerbServer(VerbDeck.instance),
      DevelServer()
    )
  }

  server.start()
}
