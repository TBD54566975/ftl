package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.METHOD, ElementType.ANNOTATION_TYPE })
public @interface SubscriptionOptions {
    /**
     *
     * @return The initial offset to start consuming from.
     */
    FromOffset from();

    /**
     *
     * @return Whether to create a dead letter queue for events that do not succeed within the retry policy.
     */
    boolean deadLetter() default false;
}