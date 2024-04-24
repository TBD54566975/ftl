package xyz.block.ftl

enum class Visibility {
  PUBLIC, INTERNAL, PRIVATE
}

enum class Ingress {
  HTTP, NONE
}

enum class Method {
  GET, POST, PUT, DELETE, NONE
}

/**
 * This annotation can be used to mark classes or functions for export with specified visibility,
 * type of ingress, HTTP method, and a routing path.
 */
@Target(AnnotationTarget.FUNCTION, AnnotationTarget.CLASS)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Export(
  val visibility: Visibility,
  val ingress: Ingress = Ingress.NONE,
  val method: Method = Method.NONE,
  val path: String = "",
)
