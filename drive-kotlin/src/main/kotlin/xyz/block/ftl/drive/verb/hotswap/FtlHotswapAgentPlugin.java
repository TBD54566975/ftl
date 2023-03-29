package xyz.block.ftl.drive.verb.hotswap;

import org.hotswap.agent.annotation.LoadEvent;
import org.hotswap.agent.annotation.OnClassLoadEvent;
import org.hotswap.agent.annotation.Plugin;
import org.hotswap.agent.javassist.CtClass;
import xyz.block.ftl.drive.Logging;

@Plugin(name = "FtlHotswapAgentPlugin", testedVersions = {})
public class FtlHotswapAgentPlugin {
  @OnClassLoadEvent(classNameRegexp = ".*", events = LoadEvent.REDEFINE)
  public static void loaded(CtClass ctClass) {
    Logging.Companion.logger(FtlHotswapAgentPlugin.class.getName())
        .info("Reloaded " + ctClass.getName());
  }
}
