package xyz.block.ftl.enums;

public class Cat implements Animal {
    public boolean isCat() {
        return true;
    }

    public boolean isDog() {
        return false;
    }

    public Cat getCat() {
        return this;
    }

    public Dog getDog() {
        return null;
    }
}
