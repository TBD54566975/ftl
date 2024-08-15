# FTL Java Runtime

This contains the code for the FTL Java runtime environment.

## Tips

### Debugging Maven commands with IntelliJ

The Java runtime is built and packaged using Maven. If you would like to debug Maven commands using Intellij:

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