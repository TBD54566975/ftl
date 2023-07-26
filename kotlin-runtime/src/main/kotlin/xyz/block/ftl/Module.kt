package xyz.block.ftl

import xyz.block.ftl.v1.schema.Module
import xyz.block.ftl.v1.schema.ModuleRuntime

class ModuleBuilder {
  internal fun build(name: String): Module {
    return Module(
      name = name,
      runtime = ModuleRuntime(language = "kotlin")
    )
  }
}

// Block function to initialise the module.
fun module(name: String, init: (ModuleBuilder.() -> Unit)?): Module {
  println("Hello from Kotlin!")
  var builder = ModuleBuilder()
  if (init != null) init(builder)
  val module = builder.build(name)
  return module
}