package xyz.block.ftl;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Used to specify the name of a FTL element within the current module.
 *
 * For elements outside the module use
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface Name {
    String value();
}
