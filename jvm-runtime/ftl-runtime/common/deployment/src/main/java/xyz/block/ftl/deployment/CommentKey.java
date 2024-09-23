package xyz.block.ftl.deployment;

public class CommentKey {
    public static String ofVerb(String verb) {
        return "verb." + verb;
    }

    public static String ofData(String data) {
        return "data." + data;
    }

    public static String ofEnum(String enumName) {
        return "enum." + enumName;
    }

    public static String ofConfig(String config) {
        return "config." + config;
    }

    public static String ofSecret(String secret) {
        return "secret." + secret;
    }
}
