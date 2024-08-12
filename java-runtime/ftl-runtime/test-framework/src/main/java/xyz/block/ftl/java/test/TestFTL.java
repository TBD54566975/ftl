package xyz.block.ftl.java.test;

import xyz.block.ftl.VerbClient;

public class TestFTL {

    public static TestFTL FTL = new TestFTL();

    public static TestFTL ftl() {
        return FTL;
    }

    public  TestFTL setSecret(String secret, byte[] value) {

        return this;
    }


}
