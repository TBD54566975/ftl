install-jars: && install-generator-jar install-runtime-jar
  mvn install

install-generator-jar:
  mvn -pl :ftl-generator install

install-runtime-jar:
  mvn -pl :ftl-runtime install