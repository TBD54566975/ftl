package xyz.block.ftl

/**
 * Ignore a class or method when registering verbs.
 */
@Target(AnnotationTarget.CLASS, AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Ignore()
