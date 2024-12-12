package xyz.block.ftl.runtime;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.Properties;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;

public class DevModeRunnerDetails implements RunnerDetails {

    private final Path path;
    private volatile Map<String, String> databases;
    private volatile String proxyAddress;
    private volatile String deployment;
    private volatile boolean closed;
    private static final Pattern dbNames = Pattern.compile("database\\.([a-zA-Z0-9]+).url");
    private volatile boolean loaded = false;

    public DevModeRunnerDetails(Path path) {
        this.path = path;
        startWatchThread();
    }

    void startWatchThread() {
        Thread watchThread = new Thread(() -> {
            while (!closed) {
                try {
                    Thread.sleep(100);
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
                if (Files.exists(path)) {
                    Properties p = new Properties();
                    try (InputStream stream = Files.newInputStream(path)) {
                        p.load(stream);
                        synchronized (this) {
                            proxyAddress = p.getProperty("proxy.bind.address");
                            deployment = p.getProperty("deployment");
                            var dbs = new HashMap<String, String>();
                            for (var addr : p.stringPropertyNames()) {
                                Matcher m = dbNames.matcher(addr);
                                if (m.matches()) {
                                    dbs.put(m.group(1), p.getProperty(addr));
                                }
                            }
                            databases = dbs;
                            loaded = true;
                            notifyAll();
                            return;
                        }

                    } catch (IOException e) {
                        throw new RuntimeException(e);
                    }
                }
            }
        });
        watchThread.setDaemon(true);
        watchThread.start();
    }

    @Override
    public String getProxyAddress() {
        waitForLoad();
        if (closed) {
            return null;
        }
        return proxyAddress;
    }

    private void waitForLoad() {
        while (proxyAddress == null && !closed) {
            synchronized (this) {
                if (proxyAddress == null && !closed) {
                    try {
                        wait();
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        throw new RuntimeException(e);
                    }
                }
            }
        }
    }

    @Override
    public Optional<DatasourceDetails> getDatabase(String database, GetDeploymentContextResponse.DbType type) {
        waitForLoad();
        if (closed) {
            return Optional.empty();
        }
        String connectionString = databases.get(database);
        if (connectionString == null) {
            return Optional.empty();
        }
        return Optional.of(new DatasourceDetails(connectionString, "ftl", "ftl"));
    }

    @Override
    public String getDeploymentKey() {
        waitForLoad();
        if (closed) {
            return null;
        }
        return deployment;
    }

    @Override
    public void close() {
        closed = true;
    }
}
