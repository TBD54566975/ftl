package xyz.block.ftl.runtime;


import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class VerbRegistry {

    private final Map<Key, VerbHandler> verbs = new ConcurrentHashMap<>();


    private record Key (String module, String name){

    }
}
