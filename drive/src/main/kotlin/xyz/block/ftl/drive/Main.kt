package xyz.block.ftl.drive

import org.eclipse.jetty.server.Server
import org.eclipse.jetty.server.ServerConnector
import org.eclipse.jetty.servlet.ServletHandler

fun main(args: Array<String>) {
  println("Starting FTL Drive...")

  val server = Server()
  server.connectors = arrayOf(ServerConnector(server).apply {
    port = 8080
  })
  server.handler = ServletHandler().apply {
    addServletWithMapping(DriveServlet::class.java, "/")
  }
  server.start()
}