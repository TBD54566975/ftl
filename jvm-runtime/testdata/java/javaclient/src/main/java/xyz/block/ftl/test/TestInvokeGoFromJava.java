package xyz.block.ftl.test;

import java.time.ZonedDateTime;
import java.util.List;
import java.util.Map;

import org.jetbrains.annotations.NotNull;

import ftl.gomodule.Animal;
import ftl.gomodule.BoolVerbClient;
import ftl.gomodule.BytesVerbClient;
import ftl.gomodule.ColorInt;
import ftl.gomodule.EmptyVerbClient;
import ftl.gomodule.ErrorEmptyVerbClient;
import ftl.gomodule.ExternalTypeVerbClient;
import ftl.gomodule.FloatVerbClient;
import ftl.gomodule.IntVerbClient;
import ftl.gomodule.NoValueTypeEnumVerbClient;
import ftl.gomodule.ObjectArrayVerbClient;
import ftl.gomodule.ObjectMapVerbClient;
import ftl.gomodule.OptionalBoolVerbClient;
import ftl.gomodule.OptionalBytesVerbClient;
import ftl.gomodule.OptionalFloatVerbClient;
import ftl.gomodule.OptionalIntVerbClient;
import ftl.gomodule.OptionalStringArrayVerbClient;
import ftl.gomodule.OptionalStringMapVerbClient;
import ftl.gomodule.OptionalStringVerbClient;
import ftl.gomodule.OptionalTestObjectOptionalFieldsVerbClient;
import ftl.gomodule.OptionalTestObjectVerbClient;
import ftl.gomodule.OptionalTimeVerbClient;
import ftl.gomodule.ParameterizedObjectVerbClient;
import ftl.gomodule.ParameterizedType;
import ftl.gomodule.Scalar;
import ftl.gomodule.SinkVerbClient;
import ftl.gomodule.SourceVerbClient;
import ftl.gomodule.StringArrayVerbClient;
import ftl.gomodule.StringList;
import ftl.gomodule.StringMapVerbClient;
import ftl.gomodule.StringVerbClient;
import ftl.gomodule.TestObject;
import ftl.gomodule.TestObjectOptionalFields;
import ftl.gomodule.TestObjectOptionalFieldsVerbClient;
import ftl.gomodule.TestObjectVerbClient;
import ftl.gomodule.TimeVerbClient;
import ftl.gomodule.TypeEnum;
import ftl.gomodule.TypeEnumVerbClient;
import ftl.gomodule.ValueEnumVerbClient;
import web5.sdk.dids.didcore.Did;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class TestInvokeGoFromJava {

    @Export
    @Verb
    public void emptyVerb(EmptyVerbClient emptyVerbClient) {
        emptyVerbClient.call();
    }

    @Export
    @Verb
    public void sinkVerb(String input, SinkVerbClient sinkVerbClient) {
        sinkVerbClient.call(input);
    }

    @Export
    @Verb
    public String sourceVerb(SourceVerbClient sourceVerbClient) {
        return sourceVerbClient.call();
    }

    /**
     * JAVA COMMENT
     */
    @Export
    @Verb
    public void errorEmptyVerb(ErrorEmptyVerbClient client) {
        client.call();
    }

    @Export
    @Verb
    public long intVerb(long val, IntVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public double floatVerb(double val, FloatVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull String stringVerb(@NotNull String val, StringVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public byte[] bytesVerb(byte[] val, BytesVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public boolean boolVerb(boolean val, BoolVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull List<String> stringArrayVerb(@NotNull List<String> val, StringArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, String> stringMapVerb(@NotNull Map<String, String> val, StringMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, TestObject> objectMapVerb(@NotNull Map<String, TestObject> val, ObjectMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull List<TestObject> objectArrayVerb(@NotNull List<TestObject> val, ObjectArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull ParameterizedType<String> parameterizedObjectVerb(@NotNull ParameterizedType<String> val,
            ParameterizedObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull ZonedDateTime timeVerb(@NotNull ZonedDateTime instant, TimeVerbClient client) {
        return client.call(instant);
    }

    @Export
    @Verb
    public @NotNull TestObject testObjectVerb(@NotNull TestObject val, TestObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull TestObjectOptionalFields testObjectOptionalFieldsVerb(@NotNull TestObjectOptionalFields val,
            TestObjectOptionalFieldsVerbClient client) {
        return client.call(val);
    }

    // now the same again but with option return / input types

    @Export
    @Verb
    public Long optionalIntVerb(Long val, OptionalIntVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Double optionalFloatVerb(Double val, OptionalFloatVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public String optionalStringVerb(String val, OptionalStringVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public byte[] optionalBytesVerb(byte[] val, OptionalBytesVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public boolean optionalBoolVerb(boolean val, OptionalBoolVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public List<String> optionalStringArrayVerb(List<String> val, OptionalStringArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Map<String, String> optionalStringMapVerb(Map<String, String> val, OptionalStringMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public ZonedDateTime optionalTimeVerb(ZonedDateTime instant, OptionalTimeVerbClient client) {
        return client.call(instant);
    }

    @Export
    @Verb
    public TestObject optionalTestObjectVerb(TestObject val, OptionalTestObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public TestObjectOptionalFields optionalTestObjectOptionalFieldsVerb(TestObjectOptionalFields val,
            OptionalTestObjectOptionalFieldsVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Did externalTypeVerb(Did val, ExternalTypeVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Animal noValueTypeEnumVerb(Animal animal, NoValueTypeEnumVerbClient client) {
        if (animal.isCat()) {
            return client.call(animal.getCat());
        } else {
            return client.call(animal.getDog());
        }
    }

    @Export
    @Verb
    public ColorInt valueEnumVerb(ColorInt color, ValueEnumVerbClient client) {
        return client.call(ColorInt.Red);
    }

    @Export
    @Verb
    public TypeEnum typeEnumVerb(TypeEnum value, TypeEnumVerbClient client) {
        if (value.isScalar()) {
            return client.call(new StringList(List.of("a", "b", "c")));
        } else if (value.isStringList()) {
            return client.call(new Scalar("scalar"));
        } else {
            throw new IllegalArgumentException("unexpected value");
        }
    }

    //    @Export
    //    @Verb
    //    public Mixed mixedEnumVerb(Mixed mixed, MixedEnumVerbClient client) {
    //        return client.call(mixed);
    //    }
}
