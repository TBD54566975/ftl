package xyz.block.ftl

import jakarta.servlet.http.HttpServletRequest
import xyz.block.ftl.drive.verb.VerbDeck

data class HttpSubContext(val headers: Map<String, String>)

data class TraceSubContext(val verbsTransited: List<VerbDeck.VerbId>)

class Context(
  val trace: TraceSubContext = TraceSubContext(listOf()),
  val http: HttpSubContext?) {

  companion object {
    fun fromHttpRequest(request: HttpServletRequest): Context {
      val headers = mutableMapOf<String, String>()
      request.headerNames.asIterator().forEach { name ->
        headers[name] = request.getHeader(name)
      }

      return Context(TraceSubContext(listOf()), HttpSubContext(headers))
    }

    fun fromLocal(propagator: Context) = Context(propagator.trace, null)
  }
}
