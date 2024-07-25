package xyz.block.ftl.logging

import ch.qos.logback.classic.Level
import ch.qos.logback.classic.Logger
import ch.qos.logback.classic.LoggerContext
import ch.qos.logback.classic.spi.ILoggingEvent
import ch.qos.logback.core.ConsoleAppender
import ch.qos.logback.core.joran.spi.ConsoleTarget
import com.fasterxml.jackson.core.JsonGenerator
import net.logstash.logback.composite.JsonProviders
import net.logstash.logback.composite.JsonWritingUtils
import net.logstash.logback.composite.loggingevent.LogLevelJsonProvider
import net.logstash.logback.composite.loggingevent.MessageJsonProvider
import net.logstash.logback.composite.loggingevent.ThrowableMessageJsonProvider
import net.logstash.logback.encoder.LoggingEventCompositeJsonEncoder
import org.slf4j.LoggerFactory
import kotlin.reflect.KClass

class Logging {
  private val lc = LoggerFactory.getILoggerFactory() as LoggerContext
  private val appender = ConsoleAppender<ILoggingEvent>()

  companion object {
    private val logging = Logging()
    private const val DEFAULT_LOG_LEVEL = "info"

    fun logger(name: String): Logger {
      val logger = logging.lc.getLogger(name) as Logger
      logger.addAppender(logging.appender)
      logger.level = Level.DEBUG
      logger.isAdditive = false /* set to true if root should log too */

      return logger
    }

    fun logger(kClass: KClass<*>): Logger {
      return logger(kClass.qualifiedName!!)
    }

    init {
      val je = LoggingEventCompositeJsonEncoder()
      je.context = logging.lc

      val providers: JsonProviders<ILoggingEvent> = je.providers
      providers.setContext(je.context)
      // Custom LogLevelJsonProvider converts level value to lowercase
      providers.addProvider(object : LogLevelJsonProvider() {
        override fun writeTo(generator: JsonGenerator, event: ILoggingEvent) {
          JsonWritingUtils.writeStringField(generator, fieldName, event.level.toString().lowercase())
        }
      })
      providers.addProvider(MessageJsonProvider())
      // Custom ThrowableMessageJsonProvider uses "error" as fieldname for throwable
      providers.addProvider(object : ThrowableMessageJsonProvider() {
        init {
          this.fieldName = "error"
        }
      })
      je.providers = providers
      je.start()

      logging.appender.target = ConsoleTarget.SystemErr.toString()
      logging.appender.context = logging.lc
      logging.appender.encoder = je
      logging.appender.start()

      val rootLogger = logger(Logger.ROOT_LOGGER_NAME)
      val rootLevelCfg = Level.valueOf(System.getenv("LOG_LEVEL") ?: DEFAULT_LOG_LEVEL)
      rootLogger.level = rootLevelCfg

      // Explicitly set log level for grpc-netty
      logger("io.grpc.netty").level = Level.WARN
    }
  }
}
