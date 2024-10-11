package xyz.block.ftl.enums;

import com.fasterxml.jackson.annotation.JsonIgnore;

import xyz.block.ftl.Enum;

/**
 * Comment on TypeEnum
 */
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
