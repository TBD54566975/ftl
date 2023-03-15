package xyz.block.ftl.drive

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.server.ServerConnector
import org.eclipse.jetty.servlet.ServletHandler
import xyz.block.ftl.control.startDrive
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

  startDrive(logger, VerbDeck.instance)
  server.start()
}
