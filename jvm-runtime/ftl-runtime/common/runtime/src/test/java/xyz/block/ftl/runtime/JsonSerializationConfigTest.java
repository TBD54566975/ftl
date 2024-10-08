package xyz.block.ftl.runtime;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import xyz.block.ftl.TypeAliasMapper;

class JsonSerializationConfigTest {

    @Test
    public void testExtraction() {
        Assertions.assertEquals(AtomicInteger.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(FullIntImplementation.class, 0));
        Assertions.assertEquals(Integer.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(FullIntImplementation.class, 1));

        Assertions.assertEquals(AtomicInteger.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(PartialIntImplementation.class, 0));
        Assertions.assertEquals(Integer.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(PartialIntImplementation.class, 1));

        Assertions.assertEquals(AtomicInteger.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(AtomicIntTypeMapping.class, 0));
        Assertions.assertEquals(Integer.class,
                JsonSerializationConfig.extractTypeAliasParamImpl(AtomicIntTypeMapping.class, 1));
    }

    @Test
    public void testTypeEnumSerialization() throws JsonProcessingException {
        JsonSerializationConfig config = new JsonSerializationConfig();
        ObjectMapper mapper = new ObjectMapper();
        config.registerTypeEnum(Animal.class, List.of(Dog.class, Cat.class));
        config.customize(mapper);

        String serializedDog = mapper.writeValueAsString(new Dog());
        Assertions.assertEquals("{\"name\":\"Dog\",\"value\":{}}", serializedDog);

        Animal animal = mapper.readValue(serializedDog, Animal.class);
        Assertions.assertTrue(animal instanceof Dog);

        String serializedCat = mapper.writeValueAsString(new Cat("Siamese", 10, "Fluffy"));
        Assertions.assertEquals("{\"name\":\"Cat\",\"value\":{\"name\":\"Fluffy\",\"breed\":\"Siamese\",\"furLength\":10}}",
                serializedCat);

        Animal cat = mapper.readValue(serializedCat, Animal.class);
        Assertions.assertTrue(cat instanceof Cat);
        Assertions.assertEquals("Fluffy", cat.getCat().getName());
    }

    @Test
    public void testValueEnumSerialization() throws JsonProcessingException {
        JsonSerializationConfig config = new JsonSerializationConfig();
        ObjectMapper mapper = new ObjectMapper();
        config.registerValueEnum(ColorInt.class);
        config.registerValueEnum(Shape.class);
        config.customize(mapper);

        String serializedRed = mapper.writeValueAsString(ColorInt.RED);
        Assertions.assertEquals("0", serializedRed);
        String serializedBlue = mapper.writeValueAsString(ColorInt.BLUE);
        Assertions.assertEquals("2", serializedBlue);

        ColorInt deserialized = mapper.readValue(serializedBlue, ColorInt.class);
        Assertions.assertEquals(ColorInt.BLUE, deserialized);

        String serializedCircle = mapper.writeValueAsString(Shape.CIRCLE);
        Assertions.assertEquals("\"circle\"", serializedCircle);

        Shape deserializedShape = mapper.readValue(serializedCircle, Shape.class);
        Assertions.assertEquals(Shape.CIRCLE, deserializedShape);
    }

    public static class AtomicIntTypeMapping implements TypeAliasMapper<AtomicInteger, Integer> {
        @Override
        public Integer encode(AtomicInteger object) {
            return object.get();
        }

        @Override
        public AtomicInteger decode(Integer serialized) {
            return new AtomicInteger(serialized);
        }
    }

    public static interface PartialIntMapper<FOO> extends TypeAliasMapper<FOO, Integer> {
    }

    public static class PartialIntImplementation implements PartialIntMapper<AtomicInteger> {
        @Override
        public Integer encode(AtomicInteger object) {
            return object.get();
        }

        @Override
        public AtomicInteger decode(Integer serialized) {
            return new AtomicInteger(serialized);
        }
    }

    public static interface FullIntMapper extends TypeAliasMapper<AtomicInteger, Integer> {
    }

    public static class FullIntImplementation implements FullIntMapper {
        @Override
        public Integer encode(AtomicInteger object) {
            return object.get();
        }

        @Override
        public AtomicInteger decode(Integer serialized) {
            return new AtomicInteger(serialized);
        }
    }

}
