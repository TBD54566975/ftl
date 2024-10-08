package xyz.block.ftl.runtime;

public class Dog implements Animal {
    public boolean isCat() {
        return false;
    }

    public boolean isDog() {
        return true;
    }

    @Override
    public Cat getCat() {
        throw new UnsupportedOperationException("Not implemented");
    }
}
