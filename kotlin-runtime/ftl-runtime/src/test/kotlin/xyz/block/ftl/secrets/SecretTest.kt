package xyz.block.ftl.secrets

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junitpioneer.jupiter.SetEnvironmentVariable

class SecretTest {
  @Test
  @SetEnvironmentVariable(key = "FTL_SECRET_SECRETS_TEST", value = "testingtesting")
  fun testSecret() {
    val secret = Secret<String>("test")
    assertEquals("testingtesting", secret.get())
  }
}
