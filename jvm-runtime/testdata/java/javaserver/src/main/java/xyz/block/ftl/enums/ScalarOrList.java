package xyz.block.ftl.enums;

import xyz.block.ftl.Enum;

@Enum
public sealed interface ScalarOrList permits Scalar, List {

}
