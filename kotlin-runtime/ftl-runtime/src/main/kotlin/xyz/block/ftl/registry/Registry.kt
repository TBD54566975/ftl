package xyz.block.ftl.registry

import io.github.classgraph.ClassGraph
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Export
import xyz.block.ftl.logging.Logging
import java.util.concurrent.ConcurrentHashMap
import kotlin.reflect.KFunction
import kotlin.reflect.jvm.javaMethod
import kotlin.reflect.jvm.kotlinFunction

const val defaultJvmModuleName = "ftl"

fun test() {}
data class Ref(val module: String, val name: String) {
  override fun toString() = "$module.$name"
}

internal fun xyz.block.ftl.v1.schema.Ref.toModel() = Ref(module, name)

/**
 * FTL module registry.
 *
 * This will contain all the Verbs that are registered in the module and will be used to dispatch requests to the
 * appropriate Verb.
 */
class Registry(val jvmModuleName: String = defaultJvmModuleName) {
  private val logger = Logging.logger(Registry::class)
  private val verbs = ConcurrentHashMap<Ref, VerbHandle<*>>()
  private var ftlModuleName: String? = null

  /** Return the FTL module name. This can only be called after one of the register* methods are called. */
  val moduleName: String
    get() {
      if (ftlModuleName == null) throw IllegalStateException("FTL module name not set, call one of the register* methods first")
      return ftlModuleName!!
    }

  /** Register all Verbs in the JVM package by walking the class graph. */
  fun registerAll() {
    logger.debug("Scanning for Verbs in ${jvmModuleName}...")
    ClassGraph()
      .enableAllInfo() // Scan classes, methods, fields, annotations
      .acceptPackages(jvmModuleName)
      .scan().use { scanResult ->
        scanResult.allClasses.flatMap {
          it.loadClass().kotlin.java.declaredMethods.asSequence()
        }.filter {
          it.isAnnotationPresent(Export::class.java) && !it.isAnnotationPresent(Ignore::class.java)
        }.forEach {
          val verb = it.kotlinFunction!!
          maybeRegisterVerb(verb)
        }
      }
  }

  val refs get() = verbs.keys.toList()

  private fun maybeRegisterVerb(function: KFunction<*>) {
    if (ftlModuleName == null) {
      ftlModuleName = ftlModuleFromJvmModule(jvmModuleName, function)
    }

    logger.debug("      @Verb ${function.name}()")
    val verbRef = Ref(module = ftlModuleName!!, name = function.name)
    val verbHandle = VerbHandle(function)
    if (verbs.containsKey(verbRef)) throw IllegalArgumentException("Duplicate Verb $verbRef")
    verbs[verbRef] = verbHandle
  }

  fun list(): Set<Ref> = verbs.keys

  /** Invoke a Verb with JSON-encoded payload and return its JSON-encoded response. */
  fun invoke(context: Context, verbRef: Ref, request: String): String {
    val verb = verbs[verbRef] ?: throw IllegalArgumentException("Unknown verb: $verbRef")
    return verb.invokeVerbInternal(context, request)
  }
}

/**
 * Return the FTL module name from a JVM module name and a top-level KFunction.
 *
 * For example, if the JVM module name is `ftl` and the qualified function name is
 * `ftl.core.foo`, then the FTL module name is `core`.
 */
fun ftlModuleFromJvmModule(jvmModuleName: String, verb: KFunction<*>): String {
  val packageName = verb.javaMethod?.declaringClass?.`package`?.name
    ?: throw IllegalArgumentException("No package for $verb")
  val qualifiedName = "$packageName.${verb.name}"
  require(qualifiedName.startsWith("$jvmModuleName.")) {
    "Function $qualifiedName must be in the form $jvmModuleName.<module>.<function>"
  }
  return qualifiedName.split(".")[1]
}
