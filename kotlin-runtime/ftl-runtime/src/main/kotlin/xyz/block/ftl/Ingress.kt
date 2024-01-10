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
annotation class Ingress(val method: Method, val path: String)
