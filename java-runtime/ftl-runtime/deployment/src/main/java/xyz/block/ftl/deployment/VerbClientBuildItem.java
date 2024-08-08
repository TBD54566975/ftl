package xyz.block.ftl.deployment;

import io.quarkus.builder.item.SimpleBuildItem;
import org.jboss.jandex.DotName;

import java.util.HashMap;
import java.util.Map;

public final class VerbClientBuildItem extends SimpleBuildItem {

    final Map<DotName, DiscoveredClients> verbClients;

    public VerbClientBuildItem(Map<DotName, DiscoveredClients> verbClients) {
        this.verbClients = new HashMap<>(verbClients);
    }

    public Map<DotName, DiscoveredClients> getVerbClients() {
        return verbClients;
    }

    public record DiscoveredClients(String name,String module, String generatedClient) {

    }
}
