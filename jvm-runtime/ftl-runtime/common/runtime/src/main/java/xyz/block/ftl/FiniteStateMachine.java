package xyz.block.ftl;

public @interface FiniteStateMachine {

    String name();

    Class<?> start();
    Class<?> end();
    Transition[] transitions();
}
