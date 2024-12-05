package xyz.block.ftl.runtime;

import java.net.URI;
import java.net.URISyntaxException;
import java.time.Duration;
<<<<<<< HEAD
import java.util.ArrayList;
import java.util.Arrays;
=======
>>>>>>> c1bb89efa (fix: hot reload endpoint uses runner proxy)
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Pattern;

import org.jboss.logging.Logger;

import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

public class FTLController implements LeaseClient {
    private static final Logger log = Logger.getLogger(FTLController.class);
    final String moduleName;
    final String deploymentName;
    private volatile FTLRunnerConnection runnerConnection;

    private static volatile FTLController controller;

    private final Map<String, GetDeploymentContextResponse.DbType> databases = new ConcurrentHashMap<>();
    private volatile Map<String, String> dynamicDatabaseAddresses = Map.of();

    /**
     * TODO: look at how init should work, this is terrible and will break dev mode
     */
    public static FTLController instance() {
        if (controller == null) {
            synchronized (FTLController.class) {
                if (controller == null) {
                    controller = new FTLController();
                }
            }
        }
        return controller;
    }

    void updateRunnerConnection(String address, Map<String, String> dynamicDatabaseAddresses) {
        var old = runnerConnection;
        this.runnerConnection = new FTLRunnerConnection(address, deploymentName, moduleName);
        this.dynamicDatabaseAddresses = dynamicDatabaseAddresses;
        old.close();
    }

    FTLController() {
        String endpoint = System.getenv("FTL_ENDPOINT");
        String ftlDeployment = System.getenv("FTL_DEPLOYMENT");
        String testEndpoint = System.getProperty("ftl.test.endpoint"); //set by the test framework
        if (testEndpoint != null) {
            endpoint = testEndpoint;
        }
        if (endpoint == null) {
            endpoint = "http://localhost:8892";
        }
        this.moduleName = System.getProperty("ftl.module.name");
        deploymentName = ftlDeployment == null ? moduleName : ftlDeployment; // We use the module name as the default deployment name for running ftl-dev
        runnerConnection = new FTLRunnerConnection(endpoint, deploymentName, moduleName);
    }

    public void registerDatabase(String name, GetDeploymentContextResponse.DbType type) {
        databases.put(name, type);
    }

    public byte[] getSecret(String secretName) {
        return runnerConnection.getSecret(secretName);
    }

    public byte[] getConfig(String config) {
        return runnerConnection.getConfig(config);
    }

    public Datasource getDatasource(String name) {
        var address = dynamicDatabaseAddresses.get(name);
        if (address != null) {
            return new Datasource(address, "ftl", "ftl");
        }
        if (databases.get(name) == GetDeploymentContextResponse.DbType.DB_TYPE_POSTGRES) {
            var proxyAddress = System.getenv("FTL_PROXY_POSTGRES_ADDRESS");
            return new Datasource("jdbc:postgresql://" + proxyAddress + "/" + name, "ftl", "ftl");
        } else if (databases.get(name) == GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL) {
            var proxyAddress = System.getenv("FTL_PROXY_MYSQL_ADDRESS_" + name.toUpperCase());
            return new Datasource("jdbc:mysql://" + proxyAddress + "/" + name, "ftl", "ftl");
        }
        List<GetDeploymentContextResponse.DSN> databasesList = runnerConnection.getDeploymentContext().getDatabasesList();
        for (var i : databasesList) {
            if (i.getName().equals(name)) {
                return Datasource.fromDSN(i.getDsn(), i.getType());
            }
        }
        return null;
    }

    public byte[] callVerb(String name, String module, byte[] payload) {
        return runnerConnection.callVerb(name, module, payload);
    }

    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {
        runnerConnection.publishEvent(topic, callingVerbName, event, key);
    }

    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        return runnerConnection.acquireLease(duration, keys);
    }

    public record Datasource(String connectionString, String username, String password) {

        public static Datasource fromDSN(String dsn, GetDeploymentContextResponse.DbType type) {
            String prefix = type.equals(GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL) ? "jdbc:mysql" : "jdbc:postgresql";
            try {
                URI uri = new URI(dsn);
                String username = "";
                String password = "";
                String userInfo = uri.getUserInfo();
                if (userInfo != null) {
                    var split = userInfo.split(":");
                    username = split[0];
                    password = split[1];
                    return new Datasource(
                            new URI(prefix, null, uri.getHost(), uri.getPort(), uri.getPath(), uri.getQuery(), null)
                                    .toASCIIString(),
                            username, password);
                } else {
                    //TODO: this is horrible, just quick hack for now
                    var matcher = Pattern.compile("[&?]user=([^?&]*)").matcher(dsn);
                    if (matcher.find()) {
                        username = matcher.group(1);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("[&?]password=([^?&]*)").matcher(dsn);
                    if (matcher.find()) {
                        password = matcher.group(1);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("^([^:]+):([^:]+)@").matcher(dsn);
                    if (matcher.find()) {
                        username = matcher.group(1);
                        password = matcher.group(2);
                        dsn = matcher.replaceAll("");
                    }
                    matcher = Pattern.compile("tcp\\(([^:)]+):([^:)]+)\\)").matcher(dsn);
                    if (matcher.find()) {
                        // Mysql has a messed up syntax
                        dsn = matcher.replaceAll(matcher.group(1) + ":" + matcher.group(2));
                    }
                    dsn = dsn.replaceAll("postgresql://", "");
                    dsn = dsn.replaceAll("postgres://", "");
                    dsn = dsn.replaceAll("mysql://", "");
                    dsn = prefix + "://" + dsn;
                    return new Datasource(dsn, username, password);
                }
            } catch (URISyntaxException e) {
                throw new RuntimeException(e);
            }
        }

    }
}
