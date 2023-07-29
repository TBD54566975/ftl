package xyz.block.ftl.registry

import io.github.classgraph.ClassGraph
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Verb
import xyz.block.ftl.logging.Logging
import java.util.concurrent.ConcurrentHashMap
import kotlin.reflect.KClass
import kotlin.reflect.KFunction
import kotlin.reflect.full.findAnnotation
import kotlin.reflect.full.hasAnnotation
import kotlin.reflect.jvm.kotlinFunction

val defaultJvmModuleName = "ftl"

data class VerbRef(val module: String, val name: String) {
  override fun toString() = "$module.$name"
}

internal fun xyz.block.ftl.v1.schema.VerbRef.toModel() = VerbRef(module, name)

/**
 * FTL module registry.
 *
 * This will contain all the Verbs that are registered in the module and will be used to dispatch requests to the
 * appropriate Verb. It is also used to generate the module schema.
 */
class Registry(val jvmModuleName: String = defaultJvmModuleName) {
  private val logger = Logging.logger(Registry::class)
  private val verbs = ConcurrentHashMap<VerbRef, VerbHandle<*>>()
  private var ftlModuleName: String? = null

  /** Register all Verbs in a class. */
  fun register(klass: KClass<out Any>) {
    var count = 0
    for (member in klass.members) {
      if (member is KFunction<*>) {
        maybeRegisterVerb(klass, member)
        count++
      }
    }
    if (count == 0) throw IllegalArgumentException("Class ${klass.qualifiedName} has no @Verb methods")
  }

  /** Register all Verbs in the JVM package by walking the class graph. */
  fun registerAll() {
    logger.info("Scanning for Verbs in ${jvmModuleName}...")
    ClassGraph()
      .enableAllInfo() // Scan classes, methods, fields, annotations
      .acceptPackages(jvmModuleName)
      .scan().use { scanResult ->
        // Use the ScanResult within the try block, e.g.
        for (clazz in scanResult.getClassesWithMethodAnnotation(Verb::class.java)) {
          val kClass = clazz.loadClass().kotlin
          if (kClass.hasAnnotation<Ignore>()) return
          clazz.methodInfo
            .filter { info -> info.hasAnnotation(Verb::class.java) && !info.hasAnnotation(Ignore::class.java) }
            .forEach { info ->
              val function = info.loadClassAndGetMethod().kotlinFunction!!
              maybeRegisterVerb(kClass, function)
            }
        }
      }
  }

  val refs get() = verbs.keys.toList()

  private fun <C : Any> maybeRegisterVerb(klass: KClass<out C>, function: KFunction<*>) {
    val verbAnnotation = function.findAnnotation<Verb>() ?: return
    val verbName = if (verbAnnotation.name == "") function.name else verbAnnotation.name
    val qualifiedName =
      klass.qualifiedName?.removePrefix("$jvmModuleName.")
        ?: throw IllegalArgumentException("Class must have a qualified name")
    val parts = qualifiedName.split(".")
    val moduleName = parts[0]
    if (parts.size < 2) {
      throw IllegalArgumentException("Class ${klass.qualifiedName} must be in the form $jvmModuleName.<module>.<class>")
    }
    ftlModuleName = moduleName

    logger.info("      @Verb ${function.name}()")
    val verbRef = VerbRef(module = ftlModuleName!!, name = verbName)
    val verbHandle = VerbHandle(klass, function)
    if (verbs.containsKey(verbRef)) throw IllegalArgumentException("Duplicate Verb $verbRef")
    verbs[verbRef] = verbHandle
  }

  fun list(): Set<VerbRef> = verbs.keys

  /** Invoke a Verb with JSON-encoded payload and return its JSON-encoded response. */
  fun invoke(context: Context, verbRef: VerbRef, request: String): String {
    val verb = verbs[verbRef] ?: throw IllegalArgumentException("Unknown verb: $verbRef")
    return verb.invokeVerbInternal(context, request)
  }
}
