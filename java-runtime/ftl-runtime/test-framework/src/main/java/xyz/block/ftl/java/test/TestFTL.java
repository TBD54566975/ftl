package xyz.block.ftl.java.test;

public class TestFTL {

    public static TestFTL FTL = new TestFTL();

    public static TestFTL ftl() {
        return FTL;
    }

    public TestFTL setSecret(String secret, byte[] value) {

        return this;
    }

}
