package {{ .Group }}

import xyz.block.ftl.Export
import xyz.block.ftl.Verb


@Export
@Verb
fun echo(req: String): String = "Hello, $req!"
