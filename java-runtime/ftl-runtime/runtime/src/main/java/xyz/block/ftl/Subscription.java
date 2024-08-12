package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
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

    /**
     * The type of the payload, if not set then it is inferred. This is mostly useful in the case where this is being
     * used as a meta annotation, as it allows the processor to easily validate that a subscription is being used correctly.
     *
     */
    Class<?> payloadType() default Object.class;

}
