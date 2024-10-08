package xyz.block.ftl.runtime;

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
}
