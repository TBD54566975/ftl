package xyz.block.ftl;

/**
 * A reference to an element, if the module name is empty it is assumed to be in the current module.
 *
 */
public @interface Ref {

    String name();

    String module() default "";
}
