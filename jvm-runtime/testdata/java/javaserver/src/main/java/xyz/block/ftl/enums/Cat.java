package xyz.block.ftl.enums;

import org.jetbrains.annotations.NotNull;

public class Cat implements Animal {
    private @NotNull String name;

    private @NotNull String breed;

    private long furLength;

    public Cat() {
    }

    public Cat(@NotNull String breed, long furLength, @NotNull String name) {
        this.breed = breed;
        this.furLength = furLength;
        this.name = name;
    }

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

    public @NotNull String getName() {
        return name;
    }

    public @NotNull String getBreed() {
        return breed;
    }

    public long getFurLength() {
        return furLength;
    }
}
