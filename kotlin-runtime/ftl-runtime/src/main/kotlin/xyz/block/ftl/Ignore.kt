package xyz.block.ftl

/**
 * Ignore a method when registering verbs.
 */
@Target(AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Ignore
