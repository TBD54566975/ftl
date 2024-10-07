package xyz.block.ftl.test;

import xyz.block.ftl.Config;
import xyz.block.ftl.Verb;

public class ConfigEndpoint {

    @Verb
    public String config(@Config("key") String key) {
        return key;
    }
}
