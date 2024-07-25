package xyz.block.ftl

import xyz.block.ftl.logging.Logging
import java.net.URI
import java.sql.Connection
import java.sql.DriverManager

private const val FTL_DSN_VAR_PREFIX = "FTL_POSTGRES_DSN"

/**
 * `Database` is a simple wrapper around the JDBC driver manager that provides a connection to the database specified
 * by the FTL_POSTGRES_DSN_<moduleName>_<dbName> environment variable.
 */
class Database(private val name: String) {
  private val logger = Logging.logger(Database::class)
  private val moduleName: String = Thread.currentThread().stackTrace[2]?.let {
    val components = it.className.split(".")
    require(components.first() == "ftl") {
      "Expected Database to be declared in package ftl.<module>, but was $it"
    }

    return@let components[1]
  } ?: throw IllegalStateException("Could not determine module name from Database declaration")

  fun <R> conn(block: (c: Connection) -> R): R {
    return try {
      val envVar = listOf(FTL_DSN_VAR_PREFIX, moduleName.uppercase(), name.uppercase()).joinToString("_")
      val dsn = System.getenv(envVar)
      require(dsn != null) { "missing DSN environment variable $envVar" }

      DriverManager.getConnection(dsnToJdbcUrl(dsn)).use {
        block(it)
      }
    } catch (e: Exception) {
      logger.error("Could not connect to database", e)
      throw e
    }
  }

  private fun dsnToJdbcUrl(dsn: String): String {
    val uri = URI(dsn)
    val scheme = uri.scheme ?: throw IllegalArgumentException("Missing scheme in DSN.")
    val userInfo = uri.userInfo?.split(":") ?: throw IllegalArgumentException("Missing userInfo in DSN.")
    val user = userInfo.firstOrNull() ?: throw IllegalArgumentException("Missing user in userInfo.")
    val password = if (userInfo.size > 1) userInfo[1] else ""
    val host = uri.host ?: throw IllegalArgumentException("Missing host in DSN.")
    val port = if (uri.port != -1) uri.port.toString() else throw IllegalArgumentException("Missing port in DSN.")
    val database = uri.path.trimStart('/')
    val parameters = uri.query?.replace("&", "?") ?: ""

    val jdbcScheme = when (scheme) {
      "postgres" -> "jdbc:postgresql"
      else -> throw IllegalArgumentException("Unsupported scheme: $scheme")
    }

    val jdbcUrl = "$jdbcScheme://$host:$port/$database?$parameters"
    return if (user.isNotBlank() && password.isNotBlank()) {
      "$jdbcUrl&user=$user&password=$password"
    } else {
      jdbcUrl
    }
  }
}
