package xyz.block.ftl;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * A FTL verb
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface Verb {
    boolean export() default false;
}
