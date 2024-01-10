package xyz.block.ftl

/**
 * A Verb marked as HttpIngress is used to handle raw HTTP requests.
 *
 * The request and response must be ftl.builtin.HttpRequest and ftl.builtin.HttpResponse respectively.
 */
@Target(AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class HttpIngress(val method: Method, val path: String)
