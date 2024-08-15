package xyz.block.ftl.runtime;

import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

public interface VerbInvoker {

    CallResponse handle(CallRequest in);
}
