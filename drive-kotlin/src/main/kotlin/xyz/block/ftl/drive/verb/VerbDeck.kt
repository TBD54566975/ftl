package xyz.block.ftl.drive.verb

import io.github.classgraph.ClassGraph
import xyz.block.ftl.Context
import xyz.block.ftl.Verb
import xyz.block.ftl.drive.Logging
import java.util.concurrent.ConcurrentHashMap
import kotlin.reflect.KFunction
import kotlin.reflect.KFunction1
import kotlin.reflect.jvm.kotlinFunction

class VerbDeck {
  private val logger = Logging.logger(VerbDeck::class)
  private var module: String? = null

  companion object {
    val instance = VerbDeck()

    fun init(module: String) {
      instance.init(module)
    }
  }

  data class VerbId(val qualifiedName: String)

  private val verbs = ConcurrentHashMap<VerbId, VerbCassette<out Any>>()

  fun init(module: String) {
    this.module = module
    logger.info("Scanning for Verbs in ${module}...")
    // Assign scanResult in try-with-resources
    ClassGraph() // Create a new ClassGraph instance
      .enableAllInfo() // Scan classes, methods, fields, annotations
      .acceptPackages(module) // Scan com.xyz and subpackages
      .disableJarScanning()
      .scan().use { scanResult ->                       // Perform the scan and return a ScanResult
        // Use the ScanResult within the try block, e.g.
        for (clazz in scanResult.getClassesWithMethodAnnotation(Verb::class.java)) {
          clazz.methodInfo.forEach { info ->
            logger.info("    @Verb ${info.name}")
            val function = info.loadClassAndGetMethod().kotlinFunction!!

            val verbId = toId(function)
            @Suppress("UNCHECKED_CAST")
            verbs[verbId] = VerbCassette(verbId, function as KFunction1<Any, Any>)
          }
        }
      }
  }

  fun fullyQualifiedName(id: VerbId): String = module!! + "." + id.qualifiedName

  fun list(): Set<VerbId> = verbs.keys

  fun lookup(name: String): VerbCassette<out Any>? = verbs[VerbId(name)]

  fun lookupFullyQualifiedName(name: String): VerbCassette<out Any>? {
    val prefix = module!! + "."
    if (name.startsWith(prefix)) {
      return verbs[VerbId(name.removePrefix(prefix))]
    }
    return null
  }

  fun dispatch(context: Context, verb: KFunction<*>, request: Any): Any {
    logger.debug("Local dispatch of ${verb.name} [trace: ${context.trace.verbsTransited}]")
    val verbId = toId(verb)
    return verbs[verbId]!!.dispatch(Context.fromLocal(verbId, context), request)
  }

  private fun toId(verb: KFunction<*>) = VerbId(verb.name)
}
