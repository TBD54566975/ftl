package xyz.block.ftl.config

import xyz.block.ftl.serializer.makeGson

class Config<T>(private val cls: Class<T>, val name: String) {
  private val module: String
  private val gson = makeGson()

  companion object {
    /**
     * A convenience method for creating a new Secret.
     *
     * <pre>
     *   val secret = Config.new<String>("test")
     * </pre>
     *
     */
    inline fun <reified T> new(name: String): Config<T> = Config(T::class.java, name) }

  init {
    val caller = Thread.currentThread().stackTrace[2].className
    require(caller.startsWith("ftl.") || caller.startsWith("xyz.block.ftl.config.")) { "Config must be defined in an FTL module not $caller" }
    val parts = caller.split(".")
    module = parts[parts.size - 2]
  }

  fun get(): T {
    val key = "FTL_CONFIG_${module.uppercase()}_${name.uppercase()}"
    val value = System.getenv(key) ?: throw Exception("Config key ${module}.${name} not found")
    return gson.fromJson(value, cls)
  }
}
