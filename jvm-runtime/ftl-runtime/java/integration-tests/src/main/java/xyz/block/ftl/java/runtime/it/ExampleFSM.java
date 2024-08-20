package xyz.block.ftl.java.runtime.it;

import xyz.block.ftl.FiniteStateMachine;
import xyz.block.ftl.Transition;
import xyz.block.ftl.Verb;

@FiniteStateMachine(name = "SomeMachine",
        start = ExampleFSM.State1.class,
        end = ExampleFSM.State4.class,
        transitions = {
                @Transition(start = ExampleFSM.State1.class, end = ExampleFSM.State2.class),
                @Transition(start = ExampleFSM.State2.class, end = ExampleFSM.State3.class),
                @Transition(start = ExampleFSM.State3.class, end = ExampleFSM.State4.class)
        }
)
public class ExampleFSM {

    public static class State1 {
        @Verb
        public void transition() {
        }
    }

    public static class State2 {
        @Verb
        public void transition() {
        }
    }

    public static class State3 {
        @Verb
        public void transition(MyTopic topic) {
        }
    }


    public static class State4 {
        @Verb
        public void transition() {
        }
    }

}
