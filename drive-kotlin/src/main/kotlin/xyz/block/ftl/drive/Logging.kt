package xyz.block.ftl.drive

import ch.qos.logback.classic.Level
import ch.qos.logback.classic.Logger
import ch.qos.logback.classic.LoggerContext
import ch.qos.logback.classic.encoder.PatternLayoutEncoder
import ch.qos.logback.classic.spi.ILoggingEvent
import ch.qos.logback.core.ConsoleAppender
import org.slf4j.LoggerFactory
import xyz.block.ftl.drive.verb.VerbDeck
import kotlin.reflect.KClass

class Logging {
  private val lc = LoggerFactory.getILoggerFactory() as LoggerContext
  private val appender = ConsoleAppender<ILoggingEvent>()

  companion object {
    private val logging = Logging()

    fun logger(name: String): Logger {
      val logger = LoggerFactory.getLogger(name) as Logger
      logger.addAppender(logging.appender)
      logger.level = Level.DEBUG
      logger.isAdditive = false /* set to true if root should log too */

      return logger
    }

    fun logger(kClass: KClass<VerbDeck>): Logger {
      return logger(kClass.qualifiedName!!)
    }

    fun init() {
      val ple = PatternLayoutEncoder()

      ple.pattern = "%date %level %logger{10} - %msg%n"
      ple.context = logging.lc
      ple.start()

      logging.appender.encoder = ple
      logging.appender.context = logging.lc
      logging.appender.start()

      logger("org.eclipse.jetty").level = Level.INFO
    }
  }
}
