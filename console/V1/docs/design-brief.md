# **FTL Console Design V1**

## **Use Cases:**

1. Discoverability/learning - use case is that of a large multi-team multi-service deployment where you can't fit everything in your head - think Cash scale. Having a graph with all of the services, verbs, what they call, who calls them, what all the types are...invaluable.
2. Day to day development/debugging - the timeline covers part of this use case. Digging into individual requests, scouring the logs, issuing test calls, and so on.

## **Objective:**

To design an intuitive graphical user interface, the FTL Console, that facilitates both the discoverability/learning and day-to-day development/debugging of large-scale systems developed using FTL.

## **Overview:**

The Console is the central dashboard for users, structured hierarchically:

1. **Controller:** The root interface of the dashboard.
2. **Modules:** Sub-components within the controller, each with a unique name, representing specific functionalities or services.
3. **Verbs:** Atomic units of code within each module, signifying specific actions or operations. There are self-contained, horizontally scalable, automatically instrumented, deployable function that accepts a single value and returns a single value.

## **Key Features:**

1. **System Overview:**

   - **Architecture Map:** An interactive bird's-eye view of the entire system, allowing users to understand the overall structure and dive deeper into specific modules or verbs.
   - **Interactive Elements:** Tooltips or info-bubbles providing brief descriptions or annotations when hovering over specific elements in the visual representation.

2. **Test Call Management:**

   - **Save & Manage Test Calls:** Users can save multiple test calls for future reference and switch between them easily.
   - **Resend Test Calls:** In case of errors, users can quickly resend the call with the observed signature.

3. **Log Management:**

   - **Filter & Search Logs:** Advanced filtering options and a search functionality to quickly locate specific events or anomalies.
   - **Trace Event Flow:** Visual representation of event flow through Verbs and modules, highlighting interactions and dependencies.
   - **Surface Errors:** Prominent display of errors occurring within Verbs and modules.

4. **Test Scenarios:**

   - **Create & Save Scenarios:** Define specific test scenarios based on requirements and save them for future use.
   - **Automatic Scenario Execution:** As users modify module code, associated test scenarios are automatically executed, providing real-time feedback.

5. **Request Inspector:**

   - **Detailed View:** Allows users to delve into individual requests, viewing headers, payloads, and responses in detail.
   - **Real-time Debugging Console:** Provides live feedback as users issue test calls, highlighting any errors or discrepancies.

6. **Fault Tolerance & Error Highlighting:**
   - **Annotation-driven Recovery:** Recover from errors by adjusting the Verb's annotation, allowing for automated retries and backoffs.
   - **AI-driven Error Detection:** AI orchestrators can rollback, quarantine, or overscale misbehaving Verbs based on developer intent.
   - **Idempotency Tagging:** Endpoints can be tagged as idempotent if they are, ensuring that they can be safely retried without side effects.

## **User Experience Goals:**

- **Intuitive Navigation:** Seamless navigation between controllers, modules, and verbs.
- **Visual Feedback:** Graphical representations, such as flowcharts or graphs, to aid in understanding event flows and system architecture.
- **Interactive Testing:** Modify test calls, apply different scenarios, and observe results in real-time.
- **Error Highlighting:** Immediate and prominent display of any errors or issues for quick identification and resolution.
