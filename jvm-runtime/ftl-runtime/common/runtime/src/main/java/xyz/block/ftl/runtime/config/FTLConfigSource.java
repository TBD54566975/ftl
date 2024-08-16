package xyz.block.ftl.runtime.config;

import java.net.URI;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;
import java.util.HashSet;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.eclipse.microprofile.config.spi.ConfigSource;

import xyz.block.ftl.runtime.FTLController;

public class FTLConfigSource implements ConfigSource {

    public static final String DATASOURCE_NAMES = "ftl-datasource-names.txt";

    final static String SEPARATE_SERVER = "quarkus.grpc.server.use-separate-server";
    final static String PORT = "quarkus.http.port";
    final static String HOST = "quarkus.http.host";

    final static String FTL_BIND = "FTL_BIND";

    final FTLController controller;

    private static final String DEFAULT_USER = "quarkus.datasource.username";
    private static final String DEFAULT_PASSWORD = "quarkus.datasource.password";
    private static final String DEFAULT_URL = "quarkus.datasource.jdbc.url";
    private static final Pattern USER_PATTERN = Pattern.compile("^quarkus\\.datasource\\.\"?([^.]+?)\"?.jdbc.username$");
    private static final Pattern PASSWORD_PATTERN = Pattern.compile("^quarkus\\.datasource\\.\"?([^.]+?)\"?.jdbc.password$");
    private static final Pattern URL_PATTERN = Pattern.compile("^quarkus\\.datasource\\.\"?([^.]+?)\"?.jdbc\\.url$");

    final Set<String> propertyNames;

    public FTLConfigSource(FTLController controller) {
        this.controller = controller;
        this.propertyNames = new HashSet<>(List.of(SEPARATE_SERVER, PORT, HOST));
        try (var in = Thread.currentThread().getContextClassLoader().getResourceAsStream(DATASOURCE_NAMES)) {
            String s = new String(in.readAllBytes(), StandardCharsets.UTF_8);
            for (String name : s.split("\n")) {
                if (name.isEmpty()) {
                    continue;
                }
                propertyNames.add("quarkus.datasource." + name + ".username");
                propertyNames.add("quarkus.datasource." + name + ".password");
                propertyNames.add("quarkus.datasource." + name + ".jdbc.url");
            }
        } catch (Exception e) {
            throw new RuntimeException("failed to read datasource file, this should have been generated as part of the build",
                    e);
        }
    }

    @Override
    public Set<String> getPropertyNames() {
        return propertyNames;
    }

    @Override
    public int getOrdinal() {
        return 400;
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
        if (s.startsWith("quarkus.datasource")) {
            System.out.println("prop: " + s);
            switch (s) {
                case DEFAULT_USER -> {
                    return Optional.ofNullable(controller.getDatasource("default")).map(FTLController.Datasource::username)
                            .orElse(null);
                }
                case DEFAULT_PASSWORD -> {
                    return Optional.ofNullable(controller.getDatasource("default")).map(FTLController.Datasource::password)
                            .orElse(null);
                }
                case DEFAULT_URL -> {
                    return Optional.ofNullable(controller.getDatasource("default"))
                            .map(FTLController.Datasource::connectionString)
                            .orElse(null);
                }
                //TODO: just support the default datasource for now
            }
            Matcher m = USER_PATTERN.matcher(s);
            if (m.matches()) {
                System.out.println("match: " + s);
                return Optional.ofNullable(controller.getDatasource(m.group(1))).map(FTLController.Datasource::username)
                        .orElse(null);
            }
            m = PASSWORD_PATTERN.matcher(s);
            if (m.matches()) {
                System.out.println("match: " + s);
                return Optional.ofNullable(controller.getDatasource(m.group(1))).map(FTLController.Datasource::password)
                        .orElse(null);
            }
            m = URL_PATTERN.matcher(s);
            if (m.matches()) {
                System.out.println("match: " + s);
                return Optional.ofNullable(controller.getDatasource(m.group(1))).map(FTLController.Datasource::connectionString)
                        .orElse(null);
            }
        }
        return null;
    }

    @Override
    public String getName() {
        return "FTL Config";
    }
}
