package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;

import org.jboss.jandex.DotName;

import io.quarkus.builder.item.SimpleBuildItem;

public final class VerbClientBuildItem extends SimpleBuildItem {

    final Map<DotName, DiscoveredClients> verbClients;

    public VerbClientBuildItem(Map<DotName, DiscoveredClients> verbClients) {
        this.verbClients = new HashMap<>(verbClients);
    }

    public Map<DotName, DiscoveredClients> getVerbClients() {
        return verbClients;
    }

    public record DiscoveredClients(String name, String module, String generatedClient) {

    }
}
