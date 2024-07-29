package xyz.block.ftl.hotswap

import org.hotswap.agent.annotation.LoadEvent
import org.hotswap.agent.annotation.OnClassLoadEvent
import org.hotswap.agent.annotation.Plugin
import org.hotswap.agent.javassist.CtClass
import xyz.block.ftl.logging.Logging

@Plugin(name = "FtlHotswapAgentPlugin", testedVersions = [])
object FtlHotswapAgentPlugin {
  private val logger = Logging.logger(FtlHotswapAgentPlugin::class)

  @JvmStatic
  @OnClassLoadEvent(classNameRegexp = ".*", events = [LoadEvent.REDEFINE])
  fun loaded(ctClass: CtClass) {
    logger.debug("Reloaded " + ctClass.name)
  }
}
