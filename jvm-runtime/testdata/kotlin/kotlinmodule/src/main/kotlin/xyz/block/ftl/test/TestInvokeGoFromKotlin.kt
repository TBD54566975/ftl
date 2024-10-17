package xyz.block.ftl.test

import ftl.gomodule.*
import web5.sdk.dids.didcore.Did
import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import java.time.ZonedDateTime

/**
 * KOTLIN COMMENT
 */
@Export
@Verb
fun emptyVerb(client: EmptyVerbClient) {
  client.emptyVerb()
}

@Export
@Verb
fun sinkVerb(input: String, client: SinkVerbClient) {
  client.sinkVerb(input)
}

@Export
@Verb
fun sourceVerb(client: SourceVerbClient): String {
  return client.sourceVerb()
}

@Export
@Verb
fun errorEmptyVerb(client: ErrorEmptyVerbClient) {
  client.errorEmptyVerb()
}

@Export
@Verb
fun intVerb(payload: Long, client: IntVerbClient): Long {
  return client.intVerb(payload)
}

@Export
@Verb
fun floatVerb(payload: Double, client: FloatVerbClient): Double {
  return client.floatVerb(payload)
}

@Export
@Verb
fun stringVerb(payload: String, client: StringVerbClient): String {
  return client.stringVerb(payload)
}

@Export
@Verb
fun bytesVerb(payload: ByteArray, client: BytesVerbClient): ByteArray {
  return client.bytesVerb(payload)
}

@Export
@Verb
fun boolVerb(payload: Boolean, client: BoolVerbClient): Boolean {
  return client.boolVerb(payload)
}

@Export
@Verb
fun stringArrayVerb(payload: List<String>, client: StringArrayVerbClient): List<String> {
  return client.stringArrayVerb(payload)
}

@Export
@Verb
fun stringMapVerb(payload: Map<String, String>, client: StringMapVerbClient): Map<String, String> {
  return client.stringMapVerb(payload)
}

@Export
@xyz.block.ftl.Verb
fun objectMapVerb(`val`: Map<String, TestObject>, client: ObjectMapVerbClient): Map<String, TestObject> {
  return client.objectMapVerb(`val`)
}

@Export
@xyz.block.ftl.Verb
fun objectArrayVerb(`val`: List<TestObject>, client: ObjectArrayVerbClient): List<TestObject> {
  return client.objectArrayVerb(`val`)
}

@Export
@xyz.block.ftl.Verb
fun parameterizedObjectVerb(
  `val`: ParameterizedType<String>,
  client: ParameterizedObjectVerbClient
): ParameterizedType<String> {
  return client.parameterizedObjectVerb(`val`)
}

@Export
@Verb
fun timeVerb(instant: ZonedDateTime, client: TimeVerbClient): ZonedDateTime {
  return client.timeVerb(instant)
}

@Export
@Verb
fun testObjectVerb(payload: TestObject, client: TestObjectVerbClient): TestObject {
  return client.testObjectVerb(payload)
}

@Export
@Verb
fun testObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields,
  client: TestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields {
  return client.testObjectOptionalFieldsVerb(payload)
}

// now the same again but with option return / input types
@Export
@Verb
fun optionalIntVerb(payload: Long?, client: OptionalIntVerbClient): Long? {
  return client.optionalIntVerb(payload)
}

@Export
@Verb
fun optionalFloatVerb(payload: Double?, client: OptionalFloatVerbClient): Double? {
  return client.optionalFloatVerb(payload)
}

@Export
@Verb
fun optionalStringVerb(payload: String?, client: OptionalStringVerbClient): String? {
  return client.optionalStringVerb(payload)
}

@Export
@Verb
fun optionalBytesVerb(payload: ByteArray?, client: OptionalBytesVerbClient): ByteArray? {
  return client.optionalBytesVerb(payload!!)
}

@Export
@Verb
fun optionalBoolVerb(payload: Boolean?, client: OptionalBoolVerbClient): Boolean? {
  return client.optionalBoolVerb(payload)
}

@Export
@Verb
fun optionalStringArrayVerb(payload: List<String>?, client: OptionalStringArrayVerbClient): List<String>? {
  return client.optionalStringArrayVerb(payload)
}

@Export
@Verb
fun optionalStringMapVerb(payload: Map<String, String>?, client: OptionalStringMapVerbClient): Map<String, String>? {
  return client.optionalStringMapVerb(payload)
}

@Export
@Verb
fun optionalTimeVerb(instant: ZonedDateTime?, client: OptionalTimeVerbClient): ZonedDateTime? {
  return client.optionalTimeVerb(instant!!)
}

@Export
@Verb
fun optionalTestObjectVerb(payload: TestObject?, client: OptionalTestObjectVerbClient): TestObject? {
  return client.optionalTestObjectVerb(payload!!)
}

@Export
@Verb
fun optionalTestObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields?,
  client: OptionalTestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields? {
  return client.optionalTestObjectOptionalFieldsVerb(payload!!)
}

@Export
@Verb
fun externalTypeVerb(did: Did, client: ExternalTypeVerbClient): Did {
  return client.externalTypeVerb(did)
}

@Export
@Verb
fun stringAliasedType(type: CustomSerializedType): CustomSerializedType {
  return type
}

@Export
@Verb
fun anyAliasedType(type: AnySerializedType): AnySerializedType {
  return type
}
