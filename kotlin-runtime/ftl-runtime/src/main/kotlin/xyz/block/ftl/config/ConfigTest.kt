package xyz.block.ftl.config

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junitpioneer.jupiter.SetEnvironmentVariable

class ConfigTest {
  @Test
  @SetEnvironmentVariable(key = "FTL_SECRET_SECRETS_TEST", value = "testingtesting")
  fun testSecret() {
    val secret = Secret<String>("test")
    assertEquals("testingtesting", secret.get())
  }
}
