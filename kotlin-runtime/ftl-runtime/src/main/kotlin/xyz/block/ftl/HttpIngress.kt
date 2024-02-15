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
annotation class HttpIngress(val method: Method, val path: String)

/**
 * A field marked with Json will be renamed to the specified name on ingress from external inputs.
 */
@Target(AnnotationTarget.FIELD)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Json(val name: String)
