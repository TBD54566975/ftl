package xyz.block.ftl.test

class Dog : Animal {
  override fun isCat() = false

  override fun isDog() = true

  override fun getCat(): Cat {
    throw RuntimeException("Dog is not Cat")
  }

  override fun getDog() = this
}
