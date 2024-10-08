package xyz.block.ftl.runtime;

import java.io.IOException;
import java.lang.reflect.Field;
import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;
import java.lang.reflect.TypeVariable;
import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import jakarta.enterprise.inject.Instance;
import jakarta.inject.Inject;
import jakarta.inject.Singleton;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.Version;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.SerializerProvider;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import com.fasterxml.jackson.databind.module.SimpleModule;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.fasterxml.jackson.databind.ser.std.StdSerializer;

import io.quarkus.arc.Unremovable;
import io.quarkus.jackson.ObjectMapperCustomizer;
import xyz.block.ftl.TypeAliasMapper;

/**
 * This class configures the FTL serialization
 */
@Singleton
@Unremovable
public class JsonSerializationConfig implements ObjectMapperCustomizer {

    final Iterable<TypeAliasMapper<?, ?>> instances;

    private record TypeEnumDefn<T>(Class<T> type, List<Class<?>> variants) {
    }

    final List<Class> valueEnums = new ArrayList<>();
    final List<TypeEnumDefn> typeEnums = new ArrayList<>();

    @Inject
    public JsonSerializationConfig(Instance<TypeAliasMapper<?, ?>> instances) {
        this.instances = instances;
    }

    JsonSerializationConfig() {
        this.instances = List.of();
    }

    @Override
    public void customize(ObjectMapper mapper) {
        mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
        SimpleModule module = new SimpleModule("ByteArraySerializer", new Version(1, 0, 0, ""));
        module.addSerializer(byte[].class, new ByteArraySerializer());
        module.addDeserializer(byte[].class, new ByteArrayDeserializer());
        mapper.disable(DeserializationFeature.ADJUST_DATES_TO_CONTEXT_TIME_ZONE);
        for (var i : instances) {
            var object = extractTypeAliasParam(i.getClass(), 0);
            var serialized = extractTypeAliasParam(i.getClass(), 1);
            module.addSerializer(object, new TypeAliasSerializer(object, serialized, i));
            module.addDeserializer(object, new TypeAliasDeSerializer(object, serialized, i));
        }
        for (var i : valueEnums) {
            module.addSerializer(i, new ValueEnumSerializer(i));
            module.addDeserializer(i, new ValueEnumDeserializer(i));
        }

        ObjectMapper cleanMapper = mapper.copy();
        for (var i : typeEnums) {
            module.addSerializer(i.type, new TypeEnumSerializer<>(i.type, cleanMapper));
            module.addDeserializer(i.type, new TypeEnumDeserializer<>(i.type, i.variants));
        }
        mapper.registerModule(module);
    }

    public <T extends Enum<T>> void registerValueEnum(Class enumClass) {
        valueEnums.add(enumClass);
    }

    public <T> void registerTypeEnum(Class<?> type, List<Class<?>> variants) {
        typeEnums.add(new TypeEnumDefn<>(type, variants));
    }

    static Class<?> extractTypeAliasParam(Class<?> target, int no) {
        return (Class<?>) extractTypeAliasParamImpl(target, no);
    }

    static Type extractTypeAliasParamImpl(Class<?> target, int no) {
        for (var i : target.getGenericInterfaces()) {
            if (i instanceof ParameterizedType) {
                ParameterizedType p = (ParameterizedType) i;
                if (p.getRawType().equals(TypeAliasMapper.class)) {
                    return p.getActualTypeArguments()[no];
                } else {
                    var result = extractTypeAliasParamImpl((Class<?>) p.getRawType(), no);
                    if (result instanceof Class<?>) {
                        return result;
                    } else if (result instanceof TypeVariable<?>) {
                        var params = ((Class<?>) p.getRawType()).getTypeParameters();
                        TypeVariable<?> tv = (TypeVariable<?>) result;
                        for (var j = 0; j < params.length; j++) {
                            if (params[j].getName().equals((tv).getName())) {
                                return p.getActualTypeArguments()[j];
                            }
                        }
                        return tv;
                    }
                }
            } else if (i instanceof Class<?>) {
                return extractTypeAliasParamImpl((Class<?>) i, no);
            }
        }
        throw new RuntimeException("Could not extract type params from " + target);
    }

    public static class ByteArraySerializer extends StdSerializer<byte[]> {

        public ByteArraySerializer() {
            super(byte[].class);
        }

        @Override
        public void serialize(byte[] value, com.fasterxml.jackson.core.JsonGenerator gen, SerializerProvider provider)
                throws IOException {
            gen.writeString(Base64.getEncoder().encodeToString(value));

        }
    }

    public static class ByteArrayDeserializer extends StdDeserializer<byte[]> {

        public ByteArrayDeserializer() {
            super(byte[].class);
        }

        @Override
        public byte[] deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
            JsonNode node = p.getCodec().readTree(p);
            String base64 = node.asText();
            return Base64.getDecoder().decode(base64);
        }
    }

    public static class TypeAliasDeSerializer<T, S> extends StdDeserializer<T> {

        final TypeAliasMapper<T, S> mapper;
        final Class<S> serializedType;

        public TypeAliasDeSerializer(Class<T> type, Class<S> serializedType, TypeAliasMapper<T, S> mapper) {
            super(type);
            this.mapper = mapper;
            this.serializedType = serializedType;
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
            var s = ctxt.readValue(p, serializedType);
            return mapper.decode(s);
        }
    }

    public static class TypeAliasSerializer<T, S> extends StdSerializer<T> {

        final TypeAliasMapper<T, S> mapper;
        final Class<S> serializedType;

        public TypeAliasSerializer(Class<T> type, Class<S> serializedType, TypeAliasMapper<T, S> mapper) {
            super(type);
            this.mapper = mapper;
            this.serializedType = serializedType;
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            var s = mapper.encode(value);
            gen.writeObject(s);
        }
    }

    public static class ValueEnumSerializer<T> extends StdSerializer<T> {
        private final Field valueField;

        public ValueEnumSerializer(Class<T> type) {
            super(type);
            try {
                this.valueField = type.getDeclaredField("value");
                valueField.setAccessible(true);
            } catch (NoSuchFieldException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            try {
                gen.writeObject(valueField.get(value));
            } catch (IllegalAccessException e) {
                throw new RuntimeException(e);
            }
        }
    }

    public static class ValueEnumDeserializer<T> extends StdDeserializer<T> {
        private final Map<Object, T> wireToEnum = new HashMap<>();
        private final Class<?> valueClass;

        public ValueEnumDeserializer(Class<T> type) {
            super(type);
            try {
                Field valueField = type.getDeclaredField("value");
                valueField.setAccessible(true);
                valueClass = valueField.getType();
                for (T ennum : type.getEnumConstants()) {
                    wireToEnum.put(valueField.get(ennum), ennum);
                }
            } catch (NoSuchFieldException | IllegalAccessException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
            Object wireVal = ctxt.readValue(p, valueClass);
            return wireToEnum.get(wireVal);
        }
    }

    public static class TypeEnumSerializer<T> extends StdSerializer<T> {
        private final ObjectMapper defaultMapper;

        public TypeEnumSerializer(Class<T> type, ObjectMapper mapper) {
            super(type);
            defaultMapper = mapper;
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            gen.writeStartObject();
            gen.writeStringField("name", value.getClass().getSimpleName());
            gen.writeFieldName("value");
            // Avoid infinite recursion by using a mapper without this serializer registered
            defaultMapper.writeValue(gen, value);
            gen.writeEndObject();
        }
    }

    public static class TypeEnumDeserializer<T> extends StdDeserializer<T> {
        private final Map<String, Class<?>> nameToVariant = new HashMap<>();

        public TypeEnumDeserializer(Class<T> type, List<Class<?>> variants) {
            super(type);
            for (var variant : variants) {
                nameToVariant.put(variant.getSimpleName(), variant);
            }
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
            ObjectNode wireValue = p.readValueAsTree();
            if (!wireValue.has("name") || !wireValue.has("value")) {
                throw new RuntimeException("Enum missing 'name' or 'value' fields");
            }
            String name = wireValue.get("name").asText();
            Class<?> variant = nameToVariant.get(name);
            if (variant == null) {
                throw new RuntimeException("Unknown variant " + name);
            }
            return (T) wireValue.get("value").traverse(p.getCodec()).readValueAs(variant);
        }
    }
}
