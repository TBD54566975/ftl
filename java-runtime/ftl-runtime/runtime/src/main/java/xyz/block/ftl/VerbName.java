package xyz.block.ftl;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Used to override the name of a verb
 */
@Retention(RetentionPolicy.RUNTIME)
public @interface VerbName {
    String value();
}
