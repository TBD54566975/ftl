package xyz.block.ftl.test

import com.fasterxml.jackson.annotation.JsonIgnore
import xyz.block.ftl.Enum

@Enum
interface Animal {
  @JsonIgnore
  fun isCat(): Boolean

  @JsonIgnore
  fun isDog(): Boolean

  @JsonIgnore
  fun getCat(): Cat

  @JsonIgnore
  fun getDog(): Dog
}
