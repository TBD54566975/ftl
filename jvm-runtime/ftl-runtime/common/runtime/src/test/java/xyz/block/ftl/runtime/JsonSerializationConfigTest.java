package xyz.block.ftl.runtime;

import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

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
