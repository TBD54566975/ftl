package xyz.block.ftl.generator

import com.github.ajalt.clikt.core.CliktCommand
import com.github.ajalt.clikt.parameters.options.default
import com.github.ajalt.clikt.parameters.options.option
import com.github.ajalt.clikt.parameters.options.required
import xyz.block.ftl.generator.ModuleGenerator.Companion.DEFAULT_MODULE_CLIENT_SUFFIX
import java.io.File

class Main : CliktCommand() {
  val endpoint by option(help = "FTL endpoint.").required()
  val dest by option(help = "Destination directory for generated code.").required()
  val module by option(help = "The FTL module name we're working on.").required()
  val moduleClientSuffix by option(
    help = "The suffix appended to FTL-generated client classes for other modules in this cluster."
  ).default(DEFAULT_MODULE_CLIENT_SUFFIX)

  override fun run() {
    val client = FTLClient(endpoint)
    val schema = client.getSchema()!!
    val outputDirectory = File(dest)
    outputDirectory.deleteRecursively()
    val gen = ModuleGenerator()
    gen.run(schema, outputDirectory, module, moduleClientSuffix)
  }
}

fun main(args: Array<String>) = Main().main(args)
