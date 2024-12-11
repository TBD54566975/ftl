package xyz.block.ftl.runtime;

import java.io.Closeable;
import java.util.Optional;

import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

/**
 * Details about the proxy endpoints provided by the runner
 *
 */
public interface RunnerDetails extends Closeable {

    String getProxyAddress();

    Optional<DatasourceDetails> getDatabase(String database, GetDeploymentContextResponse.DbType type);

    String getDeploymentKey();

    default void close() {
    }
}
