+++
title = "External Types"
description = "Using external types in your modules"
date = 2024-07-12T18:00:00+00:00
updated = 2024-07-12T18:00:00+00:00
draft = false
weight = 110
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

## Using external types

FTL supports the use of external types in your FTL modules. External types are types defined in other packages or modules that are not part of the FTL module.

The primary difference is that external types are not defined in the FTL schema, and therefore serialization and deserialization of these types is not handled 
by FTL. Instead, FTL relies on the runtime to handle serialization and deserialization of these types.

In some cases this feature can also be used to provide custom serialization and deserialization logic for types that are not directly supported by FTL, even
if they are defined in the same package as the FTL module.

To use an external type in your FTL module schema, declare a type alias over the external type:

{% code_selector() %}
<!-- go -->

```go
//ftl:typealias
type FtlType external.OtherType

//ftl:typealias
type FtlType2 = external.OtherType
```

The external type is widened to `Any` in the FTL schema, and the corresponding type alias will include metadata
for the runtime-specific type mapping:

```
typealias FtlType Any
  +typemap go "github.com/external.OtherType"
```

<!-- kotlin -->

```kotlin 
@TypeAlias(name = "OtherType")
class OtherTypeTypeMapper : TypeAliasMapper<OtherType, JsonNode> {
    override fun encode(`object`: OtherType): JsonNode {
        return TextNode.valueOf(`object`.value)
    }

    override fun decode(serialized: JsonNode): OtherType {
        if (serialized.isTextual) {
            return OtherType(serialized.textValue())
        }
        throw RuntimeException("Expected a textual value")
    }
}
```

In the example above the external type is widened to `Any` in the FTL schema, and the corresponding type alias will include metadata
for the runtime-specific type mapping:

```
typealias FtlType Any
  +typemap java "foo.bar.OtherType"
```

Note that for JVM languages `java` is always used as the runtime name, regardless of the actual language used.

It is also possible to map to any other valid FTL type (e.g. `String`) by use this as the second type parameter:

Users can achieve functionally equivalent behavior to using the external type directly by using the declared
alias (`FtlType`) in place of the external type in any other schema declarations (e.g. as the type of a Verb request). Direct usage of the external type in schema declarations is not supported;
instead, the type alias must be used.

```kotlin 
@TypeAlias(name = "OtherType")
class OtherTypeTypeMapper : TypeAliasMapper<OtherType, String> {
    override fun encode(other: OtherType): JsonNode {
        return other.value
    }

    override fun decode(serialized: String): OtherType {
        return OtherType(serialized.textValue())
    }
}
```

The corresponding type alias will be to a `String`, which makes the schema more useful:

```
typealias FtlType String
  +typemap kotlin "foo.bar.OtherType"
```

<!-- java -->
```java
@TypeAlias(name = "OtherType")
public class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, JsonNode> {
    @Override
    public JsonNode encode(OtherType object) {
        return TextNode.valueOf(object.getValue());
    }

    @Override
    public AnySerializedType decode(OtherType serialized) {
        if (serialized.isTextual()) {
            return new OtherType(serialized.textValue());
        }
        throw new RuntimeException("Expected a textual value");
    }
}
```

In the example above the external type is widened to `Any` in the FTL schema, and the corresponding type alias will include metadata
for the runtime-specific type mapping:

```
typealias FtlType Any
  +typemap java "foo.bar.OtherType"
```

It is also possible to map to any other valid FTL type (e.g. `String`) by use this as the second type parameter:


```java
@TypeAlias(name = "OtherType")
public class OtherTypeTypeMapper implements TypeAliasMapper<OtherType, String> {
    @Override
    public String encode(OtherType object) {
        return object.getValue();
    }

    @Override
    public String decode(OtherType serialized) {
        return new OtherType(serialized.textValue());
    }
}
```

The corresponding type alias will be to a `String`, which makes the schema more useful:

```
typealias FtlType String
  +typemap java "com.external.other.OtherType"
```
{% end %}


FTL will automatically serialize and deserialize the external type to the strong type indicated by the mapping.

## Cross-Runtime Type Mappings

FTL also provides the capability to declare type mappings for other runtimes. For instance, to include a type mapping for another language, you can 
annotate your type alias declaration as follows:


{% code_selector() %}
<!-- go -->

```go
//ftl:typealias
//ftl:typemap java "com.external.other.OtherType"
type FtlType external.OtherType
```

<!-- kotlin -->

```kotlin
@TypeAlias(
  name = "OtherType",
  languageTypeMappings = [LanguageTypeMapping(language = "go", type = "github.com/external.OtherType")]
)
```

<!-- java -->

```java
@TypeAlias(name = "OtherType", languageTypeMappings = {
        @LanguageTypeMapping(language = "go", type = "github.com/external.OtherType"),
})
...
```

{% end %}

In the FTL schema, this will appear as:

```
typealias FtlType Any
  +typemap go "github.com/external.OtherType"
  +typemap java "com.external.other.OtherType"
```

This allows FTL to decode the type properly in other languages, for seamless 
interoperability across different runtimes.