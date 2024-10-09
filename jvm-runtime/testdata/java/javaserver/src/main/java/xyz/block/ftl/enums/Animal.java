package xyz.block.ftl.enums;

import com.fasterxml.jackson.annotation.JsonIgnore;

import xyz.block.ftl.Enum;

@Enum
public interface Animal {
    @JsonIgnore
    boolean isCat();

    @JsonIgnore
    boolean isDog();

    @JsonIgnore
    Cat getCat();

    @JsonIgnore
    Dog getDog();
}
