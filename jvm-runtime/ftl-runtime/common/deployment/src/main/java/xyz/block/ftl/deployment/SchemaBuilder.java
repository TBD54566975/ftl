package xyz.block.ftl.deployment;

import java.io.IOException;
import java.io.OutputStream;
import java.lang.reflect.Modifier;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZonedDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.jboss.jandex.ArrayType;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.PrimitiveType;
import org.jetbrains.annotations.NotNull;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.databind.JsonNode;

import io.quarkus.arc.processor.DotNames;
import xyz.block.ftl.Export;
import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.v1.schema.Any;
import xyz.block.ftl.v1.schema.Array;
import xyz.block.ftl.v1.schema.Bool;
import xyz.block.ftl.v1.schema.Bytes;
import xyz.block.ftl.v1.schema.Data;
import xyz.block.ftl.v1.schema.Decl;
import xyz.block.ftl.v1.schema.Field;
import xyz.block.ftl.v1.schema.Float;
import xyz.block.ftl.v1.schema.Int;
import xyz.block.ftl.v1.schema.Metadata;
import xyz.block.ftl.v1.schema.MetadataAlias;
import xyz.block.ftl.v1.schema.Module;
import xyz.block.ftl.v1.schema.Optional;
import xyz.block.ftl.v1.schema.Ref;
import xyz.block.ftl.v1.schema.Time;
import xyz.block.ftl.v1.schema.Type;
import xyz.block.ftl.v1.schema.Unit;

public class SchemaBuilder {

    public static final String BUILTIN = "builtin";

    public static final DotName INSTANT = DotName.createSimple(Instant.class);
    public static final DotName ZONED_DATE_TIME = DotName.createSimple(ZonedDateTime.class);
    public static final DotName NOT_NULL = DotName.createSimple(NotNull.class);
    public static final DotName JSON_NODE = DotName.createSimple(JsonNode.class.getName());
    public static final DotName OFFSET_DATE_TIME = DotName.createSimple(OffsetDateTime.class.getName());
    public static final DotName GENERATED_REF = DotName.createSimple(GeneratedRef.class);
    public static final DotName EXPORT = DotName.createSimple(Export.class);

    final IndexView index;
    final Module.Builder moduleBuilder;
    final Map<TypeKey, ExistingRef> dataElements = new HashMap<>();
    final String moduleName;

    public SchemaBuilder(IndexView index, String moduleName) {
        this.index = index;
        this.moduleName = moduleName;
        this.moduleBuilder = Module.newBuilder()
                .setName(moduleName)
                .setBuiltin(false);
    }

    public Type buildType(org.jboss.jandex.Type type, boolean export) {
        switch (type.kind()) {
            case PRIMITIVE -> {
                var prim = type.asPrimitiveType();
                switch (prim.primitive()) {
                    case INT, LONG, BYTE, SHORT -> {
                        return Type.newBuilder().setInt(Int.newBuilder().build()).build();
                    }
                    case FLOAT, DOUBLE -> {
                        return Type.newBuilder().setFloat(Float.newBuilder().build()).build();
                    }
                    case BOOLEAN -> {
                        return Type.newBuilder().setBool(Bool.newBuilder().build()).build();
                    }
                    case CHAR -> {
                        return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                    }
                    default -> throw new RuntimeException("unknown primitive type: " + prim.primitive());
                }
            }
            case VOID -> {
                return Type.newBuilder().setUnit(Unit.newBuilder().build()).build();
            }
            case ARRAY -> {
                ArrayType arrayType = type.asArrayType();
                if (arrayType.componentType().kind() == org.jboss.jandex.Type.Kind.PRIMITIVE && arrayType
                        .componentType().asPrimitiveType().primitive() == PrimitiveType.Primitive.BYTE) {
                    return Type.newBuilder().setBytes(Bytes.newBuilder().build()).build();
                }
                return Type.newBuilder()
                        .setArray(Array.newBuilder().setElement(buildType(arrayType.componentType(), export)).build())
                        .build();
            }
            case CLASS -> {
                var clazz = type.asClassType();
                var info = index.getClassByName(clazz.name());

                PrimitiveType unboxed = PrimitiveType.unbox(clazz);
                if (unboxed != null) {
                    Type primitive = buildType(unboxed, export);
                    if (type.hasAnnotation(NOT_NULL)) {
                        return primitive;
                    }
                    return Type.newBuilder().setOptional(Optional.newBuilder().setType(primitive)).build();
                }
                if (info != null && info.hasDeclaredAnnotation(GENERATED_REF)) {
                    var ref = info.declaredAnnotation(GENERATED_REF);
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setName(ref.value("name").asString())
                                    .setModule(ref.value("module").asString()))
                            .build();
                }
                if (clazz.name().equals(DotName.STRING_NAME)) {
                    return Type.newBuilder().setString(xyz.block.ftl.v1.schema.String.newBuilder().build()).build();
                }
                if (clazz.name().equals(DotName.OBJECT_NAME) || clazz.name().equals(JSON_NODE)) {
                    return Type.newBuilder().setAny(Any.newBuilder().build()).build();
                }
                if (clazz.name().equals(OFFSET_DATE_TIME)) {
                    return Type.newBuilder().setTime(Time.newBuilder().build()).build();
                }
                if (clazz.name().equals(INSTANT)) {
                    return Type.newBuilder().setTime(Time.newBuilder().build()).build();
                }
                if (clazz.name().equals(ZONED_DATE_TIME)) {
                    return Type.newBuilder().setTime(Time.newBuilder().build()).build();
                }
                var existing = dataElements.get(new TypeKey(clazz.name().toString(), List.of()));
                if (existing != null) {
                    if (existing.exported() || !export || !existing.ref().getModule().equals(moduleName)) {
                        return Type.newBuilder().setRef(existing.ref()).build();
                    }
                    //bit of an edge case, we have an existing non-exported object that we need to export
                    for (var i = 0; i < moduleBuilder.getDeclsCount(); ++i) {
                        var decl = moduleBuilder.getDecls(i);
                        if (!decl.hasData()) {
                            continue;
                        }
                        if (decl.getData().getName().equals(existing.ref().getName())) {
                            moduleBuilder.setDecls(i,
                                    decl.toBuilder().setData(decl.getData().toBuilder().setExport(true)).build());
                            break;
                        }
                    }
                    return Type.newBuilder().setRef(existing.ref()).build();
                }
                Data.Builder data = Data.newBuilder();
                data.setName(clazz.name().local());
                data.setExport(type.hasAnnotation(EXPORT) || export);
                buildDataElement(data, clazz.name());
                moduleBuilder.addDecls(Decl.newBuilder().setData(data).build());
                Ref ref = Ref.newBuilder().setName(data.getName()).setModule(moduleName).build();
                dataElements.put(new TypeKey(clazz.name().toString(), List.of()),
                        new ExistingRef(ref, export || data.getExport()));
                return Type.newBuilder().setRef(ref).build();
            }
            case PARAMETERIZED_TYPE -> {
                var paramType = type.asParameterizedType();
                if (paramType.name().equals(DotName.createSimple(List.class))) {
                    return Type.newBuilder()
                            .setArray(Array.newBuilder().setElement(buildType(paramType.arguments().get(0), export)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(Map.class))) {
                    return Type.newBuilder().setMap(xyz.block.ftl.v1.schema.Map.newBuilder()
                            .setKey(buildType(paramType.arguments().get(0), export))
                            .setValue(buildType(paramType.arguments().get(0), export)))
                            .build();
                } else if (paramType.name().equals(DotNames.OPTIONAL)) {
                    //TODO: optional kinda sucks
                    return Type.newBuilder()
                            .setOptional(
                                    Optional.newBuilder().setType(buildType(paramType.arguments().get(0), export)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpRequest.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                    .addTypeParameters(buildType(paramType.arguments().get(0), export)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpResponse.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName())
                                    .addTypeParameters(buildType(paramType.arguments().get(0), export))
                                    .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder().build())))
                            .build();
                } else {
                    ClassInfo classByName = index.getClassByName(paramType.name());
                    var cb = ClassType.builder(classByName.name());
                    var main = buildType(cb.build(), export);
                    var builder = main.toBuilder();
                    var refBuilder = builder.getRef().toBuilder();

                    for (var arg : paramType.arguments()) {
                        refBuilder.addTypeParameters(buildType(arg, export));
                    }
                    builder.setRef(refBuilder);
                    return builder.build();
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(Data.Builder data, DotName className) {
        if (className == null || className.equals(DotName.OBJECT_NAME)) {
            return;
        }
        var clazz = index.getClassByName(className);
        if (clazz == null) {
            return;
        }
        //TODO: handle getters and setters properly, also Jackson annotations etc
        for (var field : clazz.fields()) {
            if (!Modifier.isStatic(field.flags())) {
                Field.Builder builder = Field.newBuilder().setName(field.name())
                        .setType(buildType(field.type(), data.getExport()));
                if (field.hasAnnotation(JsonAlias.class)) {
                    var aliases = field.annotation(JsonAlias.class);
                    if (aliases.value() != null) {
                        for (var alias : aliases.value().asStringArray()) {
                            builder.addMetadata(
                                    Metadata.newBuilder().setAlias(MetadataAlias.newBuilder().setKind(0).setAlias(alias)));
                        }
                    }
                }
                data.addFields(builder.build());
            }
        }
        buildDataElement(data, clazz.superName());
    }

    public SchemaBuilder addDecls(Decl decl) {
        moduleBuilder.addDecls(decl);
        return this;
    }

    public void writeTo(OutputStream out) throws IOException {
        moduleBuilder.build().writeTo(out);
    }

    record ExistingRef(Ref ref, boolean exported) {

    }

    private record TypeKey(String name, List<String> typeParams) {

    }

}
