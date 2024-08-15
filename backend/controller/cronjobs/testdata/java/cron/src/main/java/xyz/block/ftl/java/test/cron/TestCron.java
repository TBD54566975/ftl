package xyz.block.ftl.java.test.cron;

import io.quarkus.logging.Log;
import xyz.block.ftl.Cron;
import xyz.block.ftl.Export;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Verb;

import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;

public class TestCron {

    @Cron("* * * * * * *")
    public void cron() throws Exception {
        Files.writeString(Paths.get(System.getenv("DEST_FILE")), "Hello, world!");
    }

    @Cron("5m")
    public void fiveMinutes() {
    }

    @Cron("Sat")
    public void saturday() {

    }

}