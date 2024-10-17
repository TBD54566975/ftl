package xyz.block.ftl.test;

import java.time.ZonedDateTime;
import java.util.List;
import java.util.Map;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import ftl.gomodule.AnimalWrapper;
import ftl.gomodule.BoolVerbClient;
import ftl.gomodule.BytesVerbClient;
import ftl.gomodule.ColorWrapper;
import ftl.gomodule.EmptyVerbClient;
import ftl.gomodule.ErrorEmptyVerbClient;
import ftl.gomodule.ExternalTypeVerbClient;
import ftl.gomodule.FloatVerbClient;
import ftl.gomodule.IntVerbClient;
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
import ftl.gomodule.ShapeWrapper;
import ftl.gomodule.SinkVerbClient;
import ftl.gomodule.SourceVerbClient;
import ftl.gomodule.StringArrayVerbClient;
import ftl.gomodule.StringEnumVerbClient;
import ftl.gomodule.StringList;
import ftl.gomodule.StringMapVerbClient;
import ftl.gomodule.StringVerbClient;
import ftl.gomodule.TestObject;
import ftl.gomodule.TestObjectOptionalFields;
import ftl.gomodule.TestObjectOptionalFieldsVerbClient;
import ftl.gomodule.TestObjectVerbClient;
import ftl.gomodule.TimeVerbClient;
import ftl.gomodule.TypeEnumVerbClient;
import ftl.gomodule.TypeEnumWrapper;
import ftl.gomodule.TypeWrapperEnumVerbClient;
import ftl.gomodule.ValueEnumVerbClient;
import web5.sdk.dids.didcore.Did;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class TestInvokeGoFromJava {

    /**
     * JAVA COMMENT
     */
    @Export
    @Verb
    public void emptyVerb(EmptyVerbClient client) {
        client.emptyVerb();
    }

    @Export
    @Verb
    public void sinkVerb(String input, SinkVerbClient client) {
        client.sinkVerb(input);
    }

    @Export
    @Verb
    public String sourceVerb(SourceVerbClient client) {
        return client.sourceVerb();
    }

    @Export
    @Verb
    public void errorEmptyVerb(ErrorEmptyVerbClient client) {
        client.errorEmptyVerb();
    }

    @Export
    @Verb
    public long intVerb(long val, IntVerbClient client) {
        return client.intVerb(val);
    }

    @Export
    @Verb
    public double floatVerb(double val, FloatVerbClient client) {
        return client.floatVerb(val);
    }

    @Export
    @Verb
    public @NotNull String stringVerb(@NotNull String val, StringVerbClient client) {
        return client.stringVerb(val);
    }

    @Export
    @Verb
    public byte[] bytesVerb(byte[] val, BytesVerbClient client) {
        return client.bytesVerb(val);
    }

    @Export
    @Verb
    public boolean boolVerb(boolean val, BoolVerbClient client) {
        return client.boolVerb(val);
    }

    @Export
    @Verb
    public @NotNull List<String> stringArrayVerb(@NotNull List<String> val, StringArrayVerbClient client) {
        return client.stringArrayVerb(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, String> stringMapVerb(@NotNull Map<String, String> val, StringMapVerbClient client) {
        return client.stringMapVerb(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, TestObject> objectMapVerb(@NotNull Map<String, TestObject> val, ObjectMapVerbClient client) {
        return client.objectMapVerb(val);
    }

    @Export
    @Verb
    public @NotNull List<TestObject> objectArrayVerb(@NotNull List<TestObject> val, ObjectArrayVerbClient client) {
        return client.objectArrayVerb(val);
    }

    @Export
    @Verb
    public @NotNull ParameterizedType<String> parameterizedObjectVerb(@NotNull ParameterizedType<String> val,
            ParameterizedObjectVerbClient client) {
        return client.parameterizedObjectVerb(val);
    }

    @Export
    @Verb
    public @NotNull ZonedDateTime timeVerb(@NotNull ZonedDateTime instant, TimeVerbClient client) {
        return client.timeVerb(instant);
    }

    @Export
    @Verb
    public @NotNull TestObject testObjectVerb(@NotNull TestObject val, TestObjectVerbClient client) {
        return client.testObjectVerb(val);
    }

    @Export
    @Verb
    public @NotNull TestObjectOptionalFields testObjectOptionalFieldsVerb(@NotNull TestObjectOptionalFields val,
            TestObjectOptionalFieldsVerbClient client) {
        return client.testObjectOptionalFieldsVerb(val);
    }

    // now the same again but with option return / input types

    @Export
    @Verb
    public Long optionalIntVerb(Long val, OptionalIntVerbClient client) {
        return client.optionalIntVerb(val);
    }

    @Export
    @Verb
    public Double optionalFloatVerb(Double val, OptionalFloatVerbClient client) {
        return client.optionalFloatVerb(val);
    }

    @Export
    @Verb
    public @Nullable String optionalStringVerb(@Nullable String val, OptionalStringVerbClient client) {
        return client.optionalStringVerb(val);
    }

    @Export
    @Verb
    public byte @Nullable [] optionalBytesVerb(byte @Nullable [] val, OptionalBytesVerbClient client) {
        return client.optionalBytesVerb(val);
    }

    @Export
    @Verb
    public Boolean optionalBoolVerb(Boolean val, OptionalBoolVerbClient client) {
        return client.optionalBoolVerb(val);
    }

    @Export
    @Verb
    public @Nullable List<String> optionalStringArrayVerb(@Nullable List<String> val, OptionalStringArrayVerbClient client) {
        return client.optionalStringArrayVerb(val);
    }

    @Export
    @Verb
    public @Nullable Map<String, String> optionalStringMapVerb(@Nullable Map<String, String> val,
            OptionalStringMapVerbClient client) {
        return client.optionalStringMapVerb(val);
    }

    @Export
    @Verb
    public @Nullable ZonedDateTime optionalTimeVerb(@Nullable ZonedDateTime instant, OptionalTimeVerbClient client) {
        return client.optionalTimeVerb(instant);
    }

    @Export
    @Verb
    public @Nullable TestObject optionalTestObjectVerb(@Nullable TestObject val, OptionalTestObjectVerbClient client) {
        return client.optionalTestObjectVerb(val);
    }

    @Export
    @Verb
    public TestObjectOptionalFields optionalTestObjectOptionalFieldsVerb(TestObjectOptionalFields val,
            OptionalTestObjectOptionalFieldsVerbClient client) {
        return client.optionalTestObjectOptionalFieldsVerb(val);
    }

    @Export
    @Verb
    public Did externalTypeVerb(Did val, ExternalTypeVerbClient client) {
        return client.externalTypeVerb(val);
    }

    @Export
    @Verb
    public CustomSerializedType stringAliasedType(CustomSerializedType type) {
        return type;
    }

    @Export
    @Verb
    public AnySerializedType anyAliasedType(AnySerializedType type) {
        return type;
    }

    @Export
    @Verb
    public AnimalWrapper typeEnumVerb(AnimalWrapper animal, TypeEnumVerbClient client) {
        if (animal.getAnimal().isCat()) {
            return client.typeEnumVerb(new AnimalWrapper(animal.getAnimal().getCat()));
        } else {
            return client.typeEnumVerb(new AnimalWrapper(animal.getAnimal().getDog()));
        }
    }

    @Export
    @Verb
    public ColorWrapper valueEnumVerb(ColorWrapper color, ValueEnumVerbClient client) {
        return client.valueEnumVerb(color);
    }

    @Export
    @Verb
    public ShapeWrapper stringEnumVerb(ShapeWrapper shape, StringEnumVerbClient client) {
        return client.stringEnumVerb(shape);
    }

    @Export
    @Verb
    public TypeEnumWrapper typeWrapperEnumVerb(TypeEnumWrapper value, TypeWrapperEnumVerbClient client) {
        if (value.getType().isScalar()) {
            return client.typeWrapperEnumVerb(new TypeEnumWrapper(new StringList(List.of("a", "b", "c"))));
        } else if (value.getType().isStringList()) {
            return client.typeWrapperEnumVerb(new TypeEnumWrapper(new Scalar("scalar")));
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
