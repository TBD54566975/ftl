package xyz.block.ftl.enums;

public class Dog implements Animal {
    public boolean isCat() {
        return false;
    }

    public boolean isDog() {
        return true;
    }

    public Cat getCat() {
        return null;
    }

    public Dog getDog() {
        return this;
    }
}
