package com.example;

import ftl.origin.GetNonceClient;
import ftl.origin.GetNonceRequest;
import ftl.origin.GetNonceResponse;
import ftl.relay.AppendLogClient;
import ftl.relay.AppendLogRequest;
import io.quarkus.logging.Log;
import xyz.block.ftl.Cron;

public class Pulse {

    @Cron("1s")
    public void cron10s(GetNonceClient getNonceClient, AppendLogClient appendLogClient) throws Exception {
        GetNonceResponse nr = getNonceClient.getNonce(new GetNonceRequest());
        Log.infof("Cron job triggered, nonce %s", nr.getNonce());
        appendLogClient.appendLog(new AppendLogRequest(String.format("cron %s", nr.getNonce())));
    }

}
