package xyz.block.ftl

enum class Method {
    GET, POST, PUT, DELETE
}

@Target(AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Ingress(val method: Method, val path: String)
