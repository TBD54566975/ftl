package xyz.block.ftl.runtime;

import java.nio.file.Path;
import java.time.Duration;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import org.jboss.logging.Logger;

import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

public class FTLController implements LeaseClient {
    private static final Logger log = Logger.getLogger(FTLController.class);
    final String moduleName;
    private volatile FTLRunnerConnection runnerConnection;

    private static volatile FTLController controller;
    /**
     * The details of how to connect to the runners proxy. For dev mode this needs to be determined after startup,
     * which is why this needs to be pluggable.
     */
    private RunnerDetails runnerDetails = DefaultRunnerDetails.INSTANCE;

    private final Map<String, GetDeploymentContextResponse.DbType> databases = new ConcurrentHashMap<>();

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

    FTLController() {
        this.moduleName = System.getProperty("ftl.module.name");
    }

    public void registerDatabase(String name, GetDeploymentContextResponse.DbType type) {
        databases.put(name, type);
    }

    public byte[] getSecret(String secretName) {
        return getRunnerConnection().getSecret(secretName);
    }

    private FTLRunnerConnection getRunnerConnection() {
        if (runnerConnection == null) {
            synchronized (this) {
                if (runnerConnection == null) {
                    runnerConnection = new FTLRunnerConnection(runnerDetails.getProxyAddress(),
                            runnerDetails.getDeploymentKey(), moduleName);
                }
            }
        }
        return runnerConnection;
    }

    public void waitForDevModeStart(Path runnerInfo) {
        synchronized (this) {
            if (runnerConnection != null) {
                try {
                    runnerConnection.close();
                } catch (Exception e) {
                    log.error("Failed to close runner connection", e);
                }
                runnerConnection = null;
            }
            runnerDetails.close();
            runnerDetails = new DevModeRunnerDetails(runnerInfo);
        }

    }

    public byte[] getConfig(String config) {
        return getRunnerConnection().getConfig(config);
    }

    public DatasourceDetails getDatasource(String name) {
        GetDeploymentContextResponse.DbType type = databases.get(name);
        if (type != null) {
            var address = runnerDetails.getDatabase(name, type);
            if (address.isPresent()) {
                return address.get();
            }
        }
        List<GetDeploymentContextResponse.DSN> databasesList = getRunnerConnection().getDeploymentContext().getDatabasesList();
        for (var i : databasesList) {
            if (i.getName().equals(name)) {
                return DatasourceDetails.fromDSN(i.getDsn(), i.getType());
            }
        }
        return null;
    }

    public byte[] callVerb(String name, String module, byte[] payload) {
        return getRunnerConnection().callVerb(name, module, payload);
    }

    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {
        getRunnerConnection().publishEvent(topic, callingVerbName, event, key);
    }

    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        return getRunnerConnection().acquireLease(duration, keys);
    }

    public void loadDeploymentContext() {
        getRunnerConnection().getDeploymentContext();
    }
}
