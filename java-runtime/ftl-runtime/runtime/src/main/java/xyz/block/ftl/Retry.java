package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.METHOD })
public @interface Retry {
    int count() default 0;

    String minBackoff() default "";

    String maxBackoff() default "";

    String catchModule() default "";

    String catchVerb() default "";
}
