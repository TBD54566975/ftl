# Test Scenarios

So If I understand FTL correctly we can trace Verb A calls Verb B, which calls Verb C, and so on. So we should be able to generate test cases for each module and verb for a dev and as they work run those test ???

## Potential Test Cases:

1. **Basic Deployment**:

   - **TC1**: Deploy a module with a single verb and verify that it correctly executes and completes.
   - **TC2**: Deploy a module with multiple verbs where no verb calls another and verify each verb executes correctly.
   - **TC3**: Deploy a module with missing or incorrect configurations and verify that the system catches the error and provides a relevant error message.

2. **Inter-Verb Calls within the Same Module**:

   - **TC4**: Deploy a module where one verb calls another verb within the same module. Verify that the calling verb correctly triggers the called verb.
   - **TC5**: Deploy a module where a verb calls another verb with parameters. Verify that the called verb receives and processes the parameters correctly.
   - **TC6**: Deploy a module where a verb calls another verb but provides an incorrect schema or parameters. Verify that the error is caught and a relevant error message is provided.

3. **Inter-Verb Calls across Different Modules**:

   - **TC7**: Deploy two modules where a verb in Module A calls a verb in Module B. Verify that the interaction between the modules is correct and the called verb in Module B executes as expected.
   - **TC8**: Deploy two modules where the called verb in Module B expects parameters. Verify that the calling verb in Module A sends the correct parameters and that they are processed correctly in Module B.
   - **TC9**: Deploy two modules where the calling verb in Module A sends incorrect parameters (or schema) to a verb in Module B. Verify that the error is caught and a relevant error message is provided.
   - **TC10**: Deploy two modules with cyclical verb calls (e.g., Verb A in Module A calls Verb B in Module B, which in turn calls Verb A). Verify that such cyclic dependencies are detected and flagged.

4. **Complex Deployment Scenarios**:

   - **TC11**: Deploy a module with a complex chain of verb interactions (e.g., Verb A calls Verb B, which calls Verb C, and so on). Verify that the entire chain executes correctly and in the right order.
   - **TC12**: Deploy multiple modules with intertwined verb calls across them. Verify that all interactions are correct and there are no missed or incorrectly executed calls.

5. **Concurrency and Parallel Execution**:

   - **TC13**: Deploy two modules simultaneously where verbs from both modules are interdependent. Verify that the system correctly manages concurrency and that all verb interactions are correct.
   - **TC14**: Deploy a module with multiple independent verbs that can run in parallel. Verify that they execute concurrently and complete successfully.

6. **Error Handling in Verb Calls**:
   - **TC15**: Deploy a module where a called verb intentionally fails (e.g., simulating a cloud resource provisioning error). Verify that the calling verb or the system gracefully handles the error and provides relevant feedback.
