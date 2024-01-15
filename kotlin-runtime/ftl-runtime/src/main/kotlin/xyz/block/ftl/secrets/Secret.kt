package xyz.block.ftl.secrets

import xyz.block.ftl.serializer.makeGson

class Secret<T>(private val cls: Class<T>, private val name: String) {
  private val module: String
  private val gson = makeGson()

  companion object {
    /**
     * A convenience method for creating a new Secret.
     *
     * <pre>
     *   val secret = Secret.new<String>("test")
     * </pre>
     *
     */
    inline fun <reified T> new(name: String): Secret<T> = Secret(T::class.java, name) }

  init {
    val caller = Thread.currentThread().getStackTrace()[2].className
    require(caller.startsWith("ftl.") || caller.startsWith("xyz.block.ftl.secrets.")) { "Secrets must be defined in an FTL module not ${caller}" }
    val parts = caller.split(".")
    module = parts[parts.size - 2]
  }

  fun get(): T {
    val key = "FTL_SECRET_${module.uppercase()}_${name.uppercase()}"
    val value = System.getenv(key) ?: throw Exception("Secret ${module}.${name} not found")
    return gson.fromJson(value, cls)
  }
}
