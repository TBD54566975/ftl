package xyz.block.ftl.drive.verb.probe

import com.google.gson.Gson
import xyz.block.ftl.TraceSubContext
import xyz.block.ftl.drive.verb.VerbCassette

class ArgumentTracingProbe : TracingProbe {
  private val gson = Gson()
  override fun probe(cassette: VerbCassette<*>, argument: Any, context: TraceSubContext) {
    // Deep copy
    context.verbsTraced[cassette.verbId] = gson.toJson(argument)
  }

  override fun toString() = ArgumentTracingProbe::class.simpleName!!
}
