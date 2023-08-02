package xyz.block.ftl

class Context {
  fun <Req, Resp> call(verb: (Context, Req) -> Resp, req: Req): Resp {
    return verb(this, req)
  }

  inline fun <reified Cls, Req, Resp> call(verb: (Cls, Context, Req) -> Resp, req: Req): Resp {
    val constructor = Cls::class.constructors.first()
    return verb(constructor.call(), this, req)
  }
}
