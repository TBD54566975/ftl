package xyz.block.ftl;

import java.time.Duration;

/**
 * Client that can be used to acquire a FTL lease. If the lease cannot be acquired a {@link LeaseFailedException} is thrown.
 */
public interface LeaseClient {

    void acquireLease(Duration duration, String... keys) throws LeaseFailedException;
}
