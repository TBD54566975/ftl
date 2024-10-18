package xyz.block.ftl.test

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
}
