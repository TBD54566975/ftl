package xyz.block.ftl


/**
 * A field marked with Json will be renamed to the specified name on ingress from external inputs.
 */
@Target(AnnotationTarget.FIELD)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Json(val name: String)
