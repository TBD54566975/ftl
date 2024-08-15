package xyz.block.ftl;

/**
 * Checked exception that is thrown when a lease cannot be acquired
 */
public class LeaseFailedException extends Exception {

    public LeaseFailedException() {
    }

    public LeaseFailedException(String message) {
        super(message);
    }

    public LeaseFailedException(String message, Throwable cause) {
        super(message, cause);
    }

    public LeaseFailedException(Throwable cause) {
        super(cause);
    }

    public LeaseFailedException(String message, Throwable cause, boolean enableSuppression, boolean writableStackTrace) {
        super(message, cause, enableSuppression, writableStackTrace);
    }
}
