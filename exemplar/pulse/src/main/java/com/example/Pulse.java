package com.example;

import java.io.FileWriter;
import java.io.IOException;

import ftl.origin.GetNonceClient;
import ftl.origin.GetNonceRequest;
import ftl.origin.GetNonceResponse;
import io.quarkus.logging.Log;
import xyz.block.ftl.Config;
import xyz.block.ftl.Cron;

public class Pulse {

    @Cron("1s")
    public void cron(GetNonceClient getNonceClient, @Config("log_file") String lf) throws Exception {
        if (lf == null || lf.isEmpty()) {
            throw new IllegalArgumentException("log_file config not set");
        }
        GetNonceResponse nr = getNonceClient.call(new GetNonceRequest());
        Log.infof("Cron job triggered, lf: %s, received nonce: %s", lf, nr);
        appendLog(lf, "cron %s", nr);
    }

    public static void appendLog(String path, String msg, Object... args) {
        try (FileWriter writer = new FileWriter(path, true)) {
            writer.write(String.format(msg + "%n", args));
        } catch (IOException e) {
            throw new RuntimeException("Error writing to log file", e);
        }
    }

}
