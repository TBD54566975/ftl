package com.example;

import java.io.FileWriter;
import java.io.IOException;

import ftl.origin.GetNonceClient;
import ftl.origin.GetNonceRequest;
import ftl.origin.GetNonceResponse;
import ftl.relay.GetLogFileClient;
import ftl.relay.GetLogFileRequest;
import ftl.relay.GetLogFileResponse;
import io.quarkus.logging.Log;
import xyz.block.ftl.Cron;

public class Pulse {

    @Cron("1s")
    public void cron10s(GetNonceClient getNonceClient, GetLogFileClient client) throws Exception {
        GetNonceResponse nr = getNonceClient.call(new GetNonceRequest());
        GetLogFileResponse lfr = client.getLogFile(new GetLogFileRequest());
        Log.infof("Cron job triggered, nonce %s, log file: %s", nr.getNonce(), lfr.getPath());
        appendLog(lfr.getPath(), "cron %s", nr.getNonce());
    }

    public static void appendLog(String path, String msg, Object... args) {
        try (FileWriter writer = new FileWriter(path, true)) {
            writer.write(String.format(msg + "%n", args));
        } catch (IOException e) {
            throw new RuntimeException("Error writing to log file", e);
        }
    }

}
