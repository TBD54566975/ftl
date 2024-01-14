# Install all JARs to local Maven repository
install-jars: install-root-jar install-generator-jar install-runtime-jar

# Install root JAR to local Maven repository
install-root-jar:
  mvn install

# Install ftl-generator JAR to local Maven repository
install-generator-jar:
  mvn -pl :ftl-generator install

# Install ftl-runtime JAR to local Maven repository
install-runtime-jar:
  mvn -pl :ftl-runtime install

# Deploy the Go time module
deploy-time:
  ftl deploy examples/time

# Deploy the Kotlin echo module
deploy-echo-kotlin:
  ftl deploy examples/echo-kotlin

regen-schema:
  bit protos/xyz/block/ftl/v1/schema/schema.proto
  bit protos/xyz/block/ftl/v1/schema/schema.pb.go
