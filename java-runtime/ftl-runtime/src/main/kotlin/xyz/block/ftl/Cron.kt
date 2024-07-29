package xyz.block.ftl

/**
 * A Verb marked as Cron will be run on a schedule.
 */
@Target(AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
@MustBeDocumented
annotation class Cron(val pattern: String)
