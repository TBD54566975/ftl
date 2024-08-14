package xyz.block.ftl.runtime.builtin;

import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonRawValue;

/**
 * TODO: should this be generated
 */
public class HttpResponse {
    private long status;
    private Map<String, List<String>> headers;
    @JsonRawValue
    private String body;
    private Throwable error;

    public long getStatus() {
        return status;
    }

    public void setStatus(long status) {
        this.status = status;
    }

    public Map<String, List<String>> getHeaders() {
        return headers;
    }

    public void setHeaders(Map<String, List<String>> headers) {
        this.headers = headers;
    }

    public String getBody() {
        return body;
    }

    public void setBody(String body) {
        this.body = body;
    }

    public Throwable getError() {
        return error;
    }

    public void setError(Throwable error) {
        this.error = error;
    }
}
