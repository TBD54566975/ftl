package xyz.block.ftl.test

import ftl.kotlinserver.StringEnumVerbClient
import xyz.block.ftl.Export
import xyz.block.ftl.Verb

class Verbs {


  @Export
  @Verb
  fun valueEnumVerb(color: ColorWrapper): ColorWrapper {
    return color
  }

  @Export
  @Verb
  fun stringEnumVerb(shape: ShapeWrapper): ShapeWrapper {
    return shape
  }

  @Export
  @Verb
  fun typeEnumVerb(animal: AnimalWrapper): AnimalWrapper {
    return animal
  }

  @Export
  @Verb
  fun localVerbCall(client: StringEnumVerbClient): ftl.kotlinserver.ShapeWrapper {
     return client.stringEnumVerb(ftl.kotlinserver.ShapeWrapper(ftl.kotlinserver.Shape.Square))
  }

}
