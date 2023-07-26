package xyz.block.ftl.registry

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import xyz.block.ftl.v1.schema.DataRef
import xyz.block.ftl.v1.schema.Metadata
import xyz.block.ftl.v1.schema.MetadataIngress
import xyz.block.ftl.v1.schema.Verb

class SchemaReflectorKtTest {
  @Test
  fun reflectSchemaFromFunc() {
    val expected = Verb(
      name = "verb",
      request = DataRef(name = "VerbRequest"),
      response = DataRef(name = "VerbResponse"),
      metadata = listOf(
        Metadata(ingress = MetadataIngress(method = "GET", path = "/test")),
      ),
    )
    val actual = reflectSchemaFromFunc(ExampleVerb::verb)
    assertEquals(expected, actual)
  }
}