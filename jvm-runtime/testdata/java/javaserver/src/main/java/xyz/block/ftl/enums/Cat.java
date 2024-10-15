package xyz.block.ftl.enums;

import org.jetbrains.annotations.NotNull;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Comment on Type Enum variant
 */
public class Cat implements Animal {
    @JsonProperty("name")
    private @NotNull String petName;

    private @NotNull String breed;

    private long furLength;

    public Cat() {
    }

    public Cat(@NotNull String breed, long furLength, @NotNull String name) {
        this.breed = breed;
        this.furLength = furLength;
        this.petName = name;
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

    public @NotNull String getPetName() {
        return petName;
    }

    public @NotNull String getBreed() {
        return breed;
    }

    public long getFurLength() {
        return furLength;
    }
}
