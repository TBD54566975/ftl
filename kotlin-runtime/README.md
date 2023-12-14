# FTL Kotlin Runtime

This contains the code for the FTL Kotlin runtime environment.

## Tips

### Debugging Maven commands with IntelliJ

The Kotlin runtime is built and packaged using Maven. If you would like to debug Maven commands using Intellij:

1. Click  `Run->Edit Configurations...` to bring up the run configurations window.

2. Hit `+` to add a new configuration and select `Remove JVM Debug`. Provide the following configurations and save:
- `Debugger Mode`: `Attach to remote JVM`
- `Host`: `localhost`
- `Port`: `8000`
- `Command line arguments for remote JVM`: `-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:8000`

3. Run any `mvn` command, substituting `mvnDebug` in place of `mvn` (e.g. `mvnDebug compile`). This command should hang in the terminal, awaiting your remote debugger invocation.
4. Select the newly created debugger from the `Run / Debug Configurations` drop-down in the top right of your IntelliJ 
window. With the debugger selected, hit the debug icon to run it. From here the `mvn` command will 
execute, stopping at any breakpoints specified.

#### Debugging Detekt
FTL uses [Detekt](https://github.com/Ozsie/detekt-maven-plugin) to perform static analysis, extracting FTL schemas
from modules written in Kotlin. Detekt is run as part of the `compile` phase of the Maven lifecycle and thus
can be invoked by running `mvn compile` from the command line when inside an FTL module. Use the above
instructions to configure a Maven debugger and debug Detekt with `mvnDebug compile`.