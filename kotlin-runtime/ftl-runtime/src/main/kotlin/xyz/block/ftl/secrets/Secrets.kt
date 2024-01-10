package xyz.block.ftl.secrets

object Secrets {
  private const val FTL_SECRETS_ENV_VAR_PREFIX = "FTL_SECRET_"

  fun get(name: String): String {
    if (!name.startsWith(FTL_SECRETS_ENV_VAR_PREFIX)) {
      throw Exception("Invalid secret name; must start with $FTL_SECRETS_ENV_VAR_PREFIX")
    }

    return try {
      System.getenv(name)
    } catch (e: Exception) {
      throw Exception("Secret $name not found")
    }
  }
}
