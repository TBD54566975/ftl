package xyz.block.ftl.java.runtime.it;

import xyz.block.ftl.FiniteStateMachine;
import xyz.block.ftl.Transition;
import xyz.block.ftl.Verb;

@FiniteStateMachine(name = "SomeMachine",
        start = "state1",
        end = "state2",
        transitions = {
                @Transition(start = "state1", end = "state2"),
                @Transition(start = "state2", end = "state3"),
                @Transition(start = "state3", end = "state4")
        }
)
public class ExampleFSM {

    @Verb
    public void state1() {
    }

    @Verb
    public void state2() {
    }

    @Verb
    public void state3(MyTopic topic) {
    }


    @Verb
    public void state4() {
    }

}
