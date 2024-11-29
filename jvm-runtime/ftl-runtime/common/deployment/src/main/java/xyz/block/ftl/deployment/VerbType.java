package xyz.block.ftl.deployment;

import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.Type;

import xyz.block.ftl.schema.v1.Verb;

public enum VerbType {
    VERB,
    SINK,
    SOURCE,
    EMPTY;

    public static VerbType of(Verb verb) {
        if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
            return EMPTY;
        } else if (verb.getRequest().hasUnit()) {
            return SOURCE;
        } else if (verb.getResponse().hasUnit()) {
            return SINK;
        } else {
            return VERB;
        }
    }

    public static VerbType of(MethodInfo call) {
        if (call.returnType().kind() == Type.Kind.VOID && call.parametersCount() == 0) {
            return VerbType.EMPTY;
        } else if (call.returnType().kind() == Type.Kind.VOID) {
            return VerbType.SINK;
        } else if (call.parametersCount() == 0) {
            return VerbType.SOURCE;
        } else {
            return VerbType.VERB;
        }
    }
}
