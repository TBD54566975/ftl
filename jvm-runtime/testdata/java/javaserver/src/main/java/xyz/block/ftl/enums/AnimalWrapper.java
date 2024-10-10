package xyz.block.ftl.enums;

import org.jetbrains.annotations.NotNull;

public class AnimalWrapper {
    private @NotNull Animal animal;

    public AnimalWrapper() {
    }

    public AnimalWrapper(@NotNull Animal animal) {
        this.animal = animal;
    }

    public AnimalWrapper setAnimal(@NotNull Animal animal) {
        this.animal = animal;
        return this;
    }

    public @NotNull Animal getAnimal() {
        return animal;
    }
}
