package xyz.block.ftl.runtime;

import java.io.IOException;
import java.util.Base64;

import jakarta.enterprise.event.Observes;
import jakarta.json.stream.JsonGenerator;

import com.fasterxml.jackson.core.JacksonException;
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

import io.quarkus.runtime.StartupEvent;

/**
 * This class configures the FTL serialization
 */
public class JsonSerializationConfig {

    void startup(@Observes StartupEvent event, ObjectMapper mapper) {
        mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);

        SimpleModule module = new SimpleModule("ByteArraySerializer", new Version(1, 0, 0, ""));
        module.addSerializer(byte[].class, new ByteArraySerializer());
        module.addDeserializer(byte[].class, new ByteArrayDeserializer());
        mapper.disable(DeserializationFeature.ADJUST_DATES_TO_CONTEXT_TIME_ZONE);
        mapper.registerModule(module);
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

}
