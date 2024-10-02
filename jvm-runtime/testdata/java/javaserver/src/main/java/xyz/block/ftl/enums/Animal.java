package xyz.block.ftl.enums;

import xyz.block.ftl.Enum;

@Enum
public interface Animal {
    public boolean isCat();

    public boolean isDog();

    public Cat getCat();

    public Dog getDog();
}
