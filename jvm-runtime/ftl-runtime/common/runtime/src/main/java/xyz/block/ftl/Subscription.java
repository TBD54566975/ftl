package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.METHOD, ElementType.ANNOTATION_TYPE })
public @interface Subscription {
    /**
     * The topic to subscribe to.
     */
    Ref value();

    /**
     *
     * @return The subscription name, defaults to {verbName}{topicModule?}{topicName}Subscription
     */
    String name() default "";

    /**
     * The class of the topic to subscribe to, which can be used in place of directly specifying the topic name and module.
     */
    Class<? extends Topic> topicClass() default Topic.class;
}
