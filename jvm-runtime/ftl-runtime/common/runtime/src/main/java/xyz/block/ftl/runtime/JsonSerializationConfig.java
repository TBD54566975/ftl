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

import org.jboss.logging.Logger;

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

    final Instance<TypeAliasMapper<?, ?>> instances;

    private record EnumCereal(Class<?> clazz, EnumDeserializer deserializer, EnumSerializer serializer) { }
    final List<EnumCereal> enumList = new ArrayList<>();

    @Inject
    public JsonSerializationConfig(Instance<TypeAliasMapper<?, ?>> instances) {
        this.instances = instances;
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
        for (var i : enumList) {
            module.addSerializer(i.clazz, i.serializer);
            module.addDeserializer(i.clazz, i.deserializer);
        }
        mapper.registerModule(module);
    }

    public <T extends Enum<T>> void registerValueEnum(Class enumClass) {
        enumList.add(new EnumCereal(enumClass, new EnumDeserializer<T>(enumClass), new EnumSerializer<T>(enumClass)));
    }

    //    public <T> void registerEnum(String module, Class<T> enumClass) {
    //        new EnumSerializer<Object>(enumClass, new ValueEnumMapper<T>(enumClass));
    //        enumList.add(enumClass);
    //    }

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

//    //    public record WireEnum(String name, Object value) {
//    //    }
//
//    public interface EnumMapper<T> {
//        //        Object serialize(T value);
//
//        T deserialize(Object value);
//    }
//
//    static class ValueEnumMapper<T> implements EnumMapper<T> {
//
//        private final Map<Object, T> lookup = new HashMap<>();
//
//        public ValueEnumMapper(Class<T> type) {
//
//            for (T ennum : type.getEnumConstants()) {
//                try {
//                    Field valueField = ennum.getClass().getDeclaredField("value");
//                    valueField.setAccessible(true);
//                    lookup.put(valueField.get(ennum), ennum);
//                } catch (NoSuchFieldException | IllegalAccessException e) {
//                    throw new RuntimeException(e);
//                }
//            }
//        }
//
//                @Override
//                public WireEnum serialize(T value) {
//                    try {
//                        Field valueField = value.getClass().getDeclaredField("value");
//                        valueField.setAccessible(true);
//                        WireEnum wireEnum = new WireEnum(type.getSimpleName(), valueField.get(value));
//                        log.warn("Value enum mapping enum " + value + " to wire " + wireEnum);
//                        return wireEnum;
//                    } catch (NoSuchFieldException | IllegalAccessException e) {
//                        throw new RuntimeException(e);
//                    }
//                }
//
//        @Override
//        public T deserialize(Object value) {
//            T ennum = lookup.get(value);
//            //            ennum.
//            //            T ennum = Enum.valueOf(type, (String) value.value);
//            log.warn("Value enum mapping wire value " + value + " to " + ennum);
//            return ennum;
//        }
//    }

    public static class EnumSerializer<T> extends StdSerializer<T> {
        private final Field valueField;

        public EnumSerializer(Class<T> type) {
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

    public static class EnumDeserializer<T> extends StdDeserializer<T> {
        private final Map<Object, T> wireToEnum = new HashMap<>();
        private final Class<?> valueClass;

        public EnumDeserializer(Class<T> type) {
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
}
