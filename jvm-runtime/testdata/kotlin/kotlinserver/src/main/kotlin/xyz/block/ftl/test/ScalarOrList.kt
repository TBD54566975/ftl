package xyz.block.ftl.test

import xyz.block.ftl.Enum
import xyz.block.ftl.EnumHolder
import kotlin.collections.List

@Enum
sealed interface ScalarOrList

@EnumHolder
class Scalar(val value: String?) : ScalarOrList

@EnumHolder
class List(val value: List<String>?) : ScalarOrList
