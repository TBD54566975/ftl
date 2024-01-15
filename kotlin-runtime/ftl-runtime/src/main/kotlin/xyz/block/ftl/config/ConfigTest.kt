package xyz.block.ftl.config

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junitpioneer.jupiter.SetEnvironmentVariable

class ConfigTest {
  @Test
  @SetEnvironmentVariable(key = "FTL_CONFIG_CONFIG_TEST", value = "testingtesting")
  fun testSecret() {
    val config = Config.new<String>("test")
    assertEquals("testingtesting", config.get())
  }
}
