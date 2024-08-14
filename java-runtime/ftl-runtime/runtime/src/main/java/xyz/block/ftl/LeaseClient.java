package xyz.block.ftl;

import java.time.Duration;

/**
 * Client that can be used to acquire a FTL lease. If the lease cannot be acquired a {@link LeaseFailedException} is thrown.
 */
public interface LeaseClient {

    /**
     * Acquire a lease for the given keys. The lease will be held for the given duration.
     *
     * @param duration The time to acquire the lease for
     * @param keys The lease keys
     * @return A handle that can be used to release the lease
     * @throws LeaseFailedException
     */
    LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException;
}
