package xyz.block.ftl.runtime;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

/**
 * Default implementation of RunnerDetails, that uses the environment variables to get the proxy address and database
 */
public class DefaultRunnerDetails implements RunnerDetails {

    static final DefaultRunnerDetails INSTANCE = new DefaultRunnerDetails();

    private final Map<String, GetDeploymentContextResponse.DbType> databases = new ConcurrentHashMap<>();

    @Override
    public String getProxyAddress() {
        String endpoint = System.getenv("FTL_ENDPOINT");
        String testEndpoint = System.getProperty("ftl.test.endpoint"); //set by the test framework
        if (testEndpoint != null) {
            endpoint = testEndpoint;
        }
        if (endpoint == null) {
            endpoint = "http://localhost:8892";
        }
        return endpoint;
    }

    @Override
    public Optional<DatasourceDetails> getDatabase(String name, GetDeploymentContextResponse.DbType type) {
        if (type == GetDeploymentContextResponse.DbType.DB_TYPE_POSTGRES) {
            var proxyAddress = System.getenv("FTL_PROXY_POSTGRES_ADDRESS");
            return Optional.of(new DatasourceDetails("jdbc:postgresql://" + proxyAddress + "/" + name, "ftl", "ftl"));
        } else if (type == GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL) {
            var proxyAddress = System.getenv("FTL_PROXY_MYSQL_ADDRESS_" + name.toUpperCase());
            return Optional.of(new DatasourceDetails("jdbc:mysql://" + proxyAddress + "/" + name, "ftl", "ftl"));
        }
        return Optional.empty();
    }

    @Override
    public String getDeploymentKey() {
        return System.getenv("FTL_DEPLOYMENT");
    }
}
