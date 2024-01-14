package xyz.block.ftl.config

import xyz.block.ftl.serializer.makeGson

class Secret<T>(val name: String) {
  val _module: String
  val _gson = makeGson()

  init {
    val caller = Thread.currentThread().stackTrace[2].className
    require(caller.startsWith("ftl.") || caller.startsWith("xyz.block.ftl.config.")) { "Config must be defined in an FTL module not $caller" }
    val parts = caller.split(".")
    _module = parts[parts.size - 2]
  }

  inline fun <reified T> get(): T {
    val key = "FTL_CONFIG_${_module.uppercase()}_${name.uppercase()}"
    val value = System.getenv(key) ?: throw Exception("Config key ${_module}.${name} not found")
    return _gson.fromJson(value, T::class.java)
  }
}
