package xyz.block.ftl

import jakarta.servlet.http.HttpServletRequest
import xyz.block.ftl.drive.verb.VerbDeck

data class HttpSubContext(
  val headers: Map<String, String>,
  val params: Map<String, String>,
  val uri: String)

data class TraceSubContext(val verbsTransited: List<VerbDeck.VerbId>)

class Context(
  val trace: TraceSubContext = TraceSubContext(listOf()),
  private val http: HttpSubContext?) {

  fun http(): HttpSubContext = http!!

  companion object {
    fun fromHttpRequest(verbId: VerbDeck.VerbId, request: HttpServletRequest): Context {
      val headers = mutableMapOf<String, String>()
      request.headerNames.asIterator().forEach { name ->
        headers[name] = request.getHeader(name)
      }

      val params = mutableMapOf<String, String>()
      request.parameterNames.asIterator().forEach { name ->
        params[name] = request.getParameter(name)
      }

      return Context(
        TraceSubContext(listOf(verbId)),
        HttpSubContext(headers.toMap(), params.toMap(), request.requestURI))
    }

    fun fromLocal(verbId: VerbDeck.VerbId, propagator: Context) = Context(
      TraceSubContext(propagator.trace.verbsTransited + verbId), null)
  }
}
