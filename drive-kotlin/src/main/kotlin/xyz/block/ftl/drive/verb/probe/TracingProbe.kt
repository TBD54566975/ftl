package xyz.block.ftl.drive.verb.probe

import xyz.block.ftl.TraceSubContext
import xyz.block.ftl.drive.verb.VerbCassette

interface TracingProbe {
  fun probe(cassette: VerbCassette<*>, argument: Any, context: TraceSubContext)
}
