package xyz.block.ftl;

public @interface FiniteStateMachine {

    String name();

    String start();
    String end();
    Transition[] transitions();
}
