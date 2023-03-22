package xyz.block.ftl

import xyz.block.ftl.drive.verb.VerbDeck
import kotlin.reflect.KFunction

enum class LatencyTier {
  LOCAL, RACK, DC, REGION
}

enum class Format {
  JSON, PROTO, BINARY_JSON, MULIPART_FORMDATA
}

class Ftl {
  companion object {
    fun <R> call(context: Context, verb: KFunction<R>, request: Any): R {
      @Suppress("UNCHECKED_CAST")
      return VerbDeck.instance.dispatch(context, verb, request) as R
    }

    fun schema(scm: FtlSchema.() -> Unit): FtlSchema {
      return FtlSchema()
    }
  }
}

class FtlSchema {
  class FtlConnectors {
    class HttpConnector {
      fun routes(vararg path: String) {}
      fun formats(vararg fmt: Format) {
      }
    }

    fun http(httpConf: HttpConnector.() -> Unit): HttpConnector {
      return HttpConnector()
    }

    fun grpc() {}
  }

  fun connectors(conns: FtlConnectors.() -> Unit) {

  }

  class FtlRuntime {
    fun autoscale() {}
    fun latencyTier(tier: LatencyTier) {

    }
  }

  fun runtime(rt: FtlRuntime.() -> Unit): FtlRuntime {
    return FtlRuntime()
  }
}
