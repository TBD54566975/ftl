package xyz.block.ftl.runtime.http;

import xyz.block.ftl.runtime.VerbInvoker;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

public class HTTPVerbInvoker implements VerbInvoker {

    /**
     * If this is true then the request is base 64 encoded bytes
     */
    final boolean base64Encoded;
    final FTLHttpHandler ftlHttpHandler;

    public HTTPVerbInvoker(boolean base64Encoded, FTLHttpHandler ftlHttpHandler) {
        this.base64Encoded = base64Encoded;
        this.ftlHttpHandler = ftlHttpHandler;
    }

    @Override
    public CallResponse handle(CallRequest in) {
        return ftlHttpHandler.handle(in, base64Encoded);
    }
}
