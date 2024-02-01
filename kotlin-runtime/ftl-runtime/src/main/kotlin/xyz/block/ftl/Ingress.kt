package xyz.block.ftl

enum class Method {
  GET, POST, PUT, DELETE
}

/**
 * A Verb marked as Ingress accepts HTTP requests, where the request is decoded into an arbitrary FTL type.
 */
@Target(AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Ingress(val method: Method, val path: String, val type: Type = Type.HTTP) {
  enum class Type {
    HTTP
  }
}

/**
 * A field marked with Alias will be renamed to the specified name on ingress from external inputs.
 */
@Target(AnnotationTarget.FIELD)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Alias(val name: String)
