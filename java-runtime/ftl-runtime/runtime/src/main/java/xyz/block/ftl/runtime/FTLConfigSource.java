package xyz.block.ftl.runtime;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

public class FTLConfigSource implements ConfigSource {

    final static String SEPARATE_SERVER = "quarkus.grpc.server.use-separate-server";
    final static String PORT = "quarkus.http.port";
    final static String HOST = "quarkus.http.host";

    final static String FTL_BIND = "FTL_BIND";

    @Override
    public Set<String> getPropertyNames() {
        return Set.of(SEPARATE_SERVER, PORT, HOST);
    }

    @Override
    public int getOrdinal() {
        return 1;
    }

    @Override
    public String getValue(String s) {
        switch (s) {
            case SEPARATE_SERVER -> {
                return "false";
            }
            case PORT -> {
                String bind = System.getenv(FTL_BIND);
                if (bind == null) {
                    return null;
                }
                try {
                    URI uri = new URI(bind);
                    return Integer.toString(uri.getPort());
                } catch (URISyntaxException e) {
                    return null;
                }
            }
            case HOST -> {
                String bind = System.getenv(FTL_BIND);
                if (bind == null) {
                    return null;
                }
                try {
                    URI uri = new URI(bind);
                    return uri.getHost();
                } catch (URISyntaxException e) {
                    return null;
                }
            }
        }
        return null;
    }

    @Override
    public String getName() {
        return "FTL Config";
    }
}
