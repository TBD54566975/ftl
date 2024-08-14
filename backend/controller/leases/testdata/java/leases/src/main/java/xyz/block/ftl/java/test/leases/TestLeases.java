package xyz.block.ftl.java.test.leases;

import io.quarkus.logging.Log;
import xyz.block.ftl.Export;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Verb;

import java.time.Duration;

public class TestLeases {

    @Export
    @Verb
    public void acquire(LeaseClient leaseClient) throws Exception {
        Log.info("Acquiring lease");
        try (var lease = leaseClient.acquireLease(Duration.ofSeconds(10), "lease")) {
            Log.info("Acquired lease");
            Thread.sleep(5000);
        }
    }

}
