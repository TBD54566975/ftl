package xyz.block.ftl

import org.junit.jupiter.api.Test
import xyz.block.ftl.v1.schema.Module
import xyz.block.ftl.v1.schema.ModuleRuntime
import kotlin.test.assertEquals

class ModuleKtTest {
  @Test
  fun testModuleName() {
    val actual = module("test") {}
    val expected = Module(name = "test", runtime = ModuleRuntime(language = "kotlin"))
    assertEquals(expected, actual)
  }
}