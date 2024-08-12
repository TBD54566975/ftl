package xyz.block.ftl;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Used to override the name of a verb. Without this annotation it defaults to the method name.
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface VerbName {
    String value();
}
