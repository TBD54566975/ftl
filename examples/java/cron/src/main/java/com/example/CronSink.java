package com.example;

import io.quarkus.logging.Log;
import xyz.block.ftl.Cron;

public class CronSink {

    @Cron("10s")
    public void cron10s() {
        Log.infof("Frequent cron job triggered");
    }

    @Cron("0 * * * * ")
    public void cronHourly() {
        Log.infof("Hourly cron job triggered");
    }
}
