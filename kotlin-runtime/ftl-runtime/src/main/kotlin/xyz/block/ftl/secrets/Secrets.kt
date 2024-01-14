package xyz.block.ftl.secrets

import xyz.block.ftl.serializer.makeGson

class Secret<T>(val name: String) {
  val module: String
  val gson = makeGson()

  init {
    val caller = Thread.currentThread().getStackTrace()[2].className
    require(caller.startsWith("ftl.") || caller.startsWith("xyz.block.ftl.secrets.")) { "Secrets must be defined in an FTL module not ${caller}" }
    val parts = caller.split(".")
    module = parts[parts.size - 2]
  }

  inline fun <reified T> get(): T {
    val key = "FTL_SECRET_${module.uppercase()}_${name.uppercase()}"
    val value = System.getenv(key) ?: throw Exception("Secret ${module}.${name} not found")
    return gson.fromJson(value, T::class.java)
  }
}
