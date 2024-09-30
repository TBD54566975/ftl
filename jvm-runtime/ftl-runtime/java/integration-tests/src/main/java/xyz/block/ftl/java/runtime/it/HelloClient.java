package xyz.block.ftl.java.runtime.it;

import xyz.block.ftl.VerbClient;
import xyz.block.ftl.VerbClientDefinition;

@VerbClientDefinition(name = "hello")
public interface HelloClient extends VerbClient<String, String> {
}