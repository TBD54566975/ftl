+++
title = "Start"
description = "Preparing to use FTL."
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 10
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

## Import the runtime

{% code_selector() %}

<!-- go -->

Some aspects of FTL rely on a runtime which must be imported with:

```go
import "github.com/block/ftl/go-runtime/ftl"
```

<!-- kotlin -->

The easiest way to get started with Kotlin is to use the `ftl-build-parent-kotlin` parent POM. This will automatically include the FTL runtime as a dependency, and setup all required plugins:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.myproject</groupId>
    <artifactId>myproject</artifactId>
    <version>1.0-SNAPSHOT</version>

    <parent>
        <groupId>xyz.block.ftl</groupId>
        <artifactId>ftl-build-parent-kotlin</artifactId>
        <version>${ftl.version}</version>
    </parent>

</project>
```

If you do not want to use a parent pom then you can copy the plugins and dependencies from the parent pom into your own pom.

<!-- java -->

The easiest way to get started with Java is to use the `ftl-build-parent-java` parent POM. This will automatically include the FTL runtime as a dependency, and setup all required plugins:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.myproject</groupId>
    <artifactId>myproject</artifactId>
    <version>1.0-SNAPSHOT</version>

    <parent>
        <groupId>xyz.block.ftl</groupId>
        <artifactId>ftl-build-parent-java</artifactId>
        <version>${ftl.version}</version>
    </parent>

</project>
```

If you do not want to use a parent pom then you can copy the plugins and dependencies from the parent pom into your own pom.

{% end %}
