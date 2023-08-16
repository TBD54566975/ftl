package xyz.block.ftl.schemaextractor

import xyz.block.ftl.registry.Registry

fun main() {
  val registry = Registry()
  registry.registerAll()
  println(registry.schema())
}
