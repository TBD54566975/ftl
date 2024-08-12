package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.METHOD, ElementType.ANNOTATION_TYPE })
public @interface Subscription {
    /**
     * @return The module of the topic to subscribe to, if empty then the topic is assumed to be in the current module.
     */
    String module() default "";

    /**
     *
     * @return The name of the topic to subscribe to.
     */
    String topic();

    /**
     *
     * @return The subscription name
     */
    String name();

}
