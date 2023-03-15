package xyz.block.ftl.drive.transport

import jakarta.servlet.http.HttpServlet
import jakarta.servlet.http.HttpServletRequest
import jakarta.servlet.http.HttpServletResponse
import jakarta.servlet.http.HttpServletResponse.SC_OK
import xyz.block.ftl.Context
import xyz.block.ftl.drive.adapter.JsonAdapter
import xyz.block.ftl.drive.verb.VerbDeck

class DriveServlet : HttpServlet() {
  private val deck = VerbDeck.instance
  private val jsonAdapter = JsonAdapter()

  override fun doPost(request: HttpServletRequest?, response: HttpServletResponse?) {
    response!!.apply {
      contentType = "application/json"
      status = SC_OK

      // Simple janky mapping between request URI and verb name
      val name = request!!.requestURI.substring(1)
      val verb = deck.lookup(name)
      checkNotNull(verb) { "No such verb available: ${name}" }

      // Use "Connectors" as a layer between http and the verb deck
      val input = jsonAdapter.readAs(request.reader, verb.argumentType)

      val output = deck.dispatch(Context.fromHttpRequest(verb.verbId, request), verb.verbId, input)

      jsonAdapter.write(output, response.writer)
    }
  }
}
