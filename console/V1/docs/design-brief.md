# **FTL Console Design V1**

## **Use Cases:**

1. **Discoverability/learning:** Imagine a large multi-team multi-service deployment where you can't fit everything in your head - think Cash scale. What we need then is to have a visualization with all of the services, verbs, what they call, who calls them, what all the types are.
2. **Day to day development/debugging:** - We also need to be able to dig into individual requests, scouring the logs, issuing test calls, and so on.

## **Objective:**

We want to design an intuitive graphical user interface, the FTL Console, that facilitates both the discoverability/learning and day-to-day development/debugging of large-scale systems developed using FTL.

## **Overview:**

The Console is our central dashboard, structured hierarchically:

1. **Controller:** The root interface of the dashboard.
2. **Modules:** Sub-components within the controller, each with a unique name, representing specific functionalities or services.
3. **Verbs:** Atomic units of code within each module, signifying specific actions or operations. There are self-contained, horizontally scalable, automatically instrumented, deployable function that accepts a single value and returns a single value. It is to these we issue our test calls.

## **User Experience Goals:**

- **Intuitive Navigation:** Seamless navigation between a controller's modules and verbs.
- **Visual Feedback:** Graphical representations, such as flowcharts or graphs, to aid in understanding event flows and system architecture.
- **Interactive Testing:** Modify test calls, apply different scenarios, and observe results in real-time.
- **Error Highlighting:** Immediate and prominent display of any errors or issues for quick identification and resolution.

## **Features:**

1.  **Timescale Management:**

    1. **Time Scrubbing:** We can move back and forth through time to see how events evolve in the console. This is particularly useful to track events across modules.
    2. **Broadcast Selected Timescale**: When we scrub the timeline this feature sends messages to other **features** indicating the ranges of time to display information for.
    3. **Easy Time Range Selection:** Select between easy time ranges, 5 minutes, 15 minutes, 30 minutes, last hour, Last unit of time.
    4. **Surface Errors:** For search param based errors it broadcast the errors to subscribing features.

2.  **Query Management:**

    1. **Send Search:** We can submit a search query based on keywords, event IDs, module names, log & log levels, calls, deployments, or specific error messages.
    2. **Chat Dialog:** When using this feature we can enter a chat dialog flow shifting from search mode. In this mode the dialog can answer questions and make subsequent queries.
    3. **Broadcast Queries**: When we make queries this feature sends messages to other **features** indicating the query parameters.

3.  **Test Call Management:**

    1. **Send Test Call:** We can input or select the necessary parameters for a test call and send it. This interface also displays the expected schema for the selected test call.
    2. **Save Test Call:** We can also save a test call with a unique name and do the following actions to it:
       1. **Rename**
       2. **Delete**
       3. **Update**
       4. **Copy as**
       5. **Add pass or fail assertion:** When we do this this will generate a test case that will run whenever the module is modified or deployed.
    3. **Surface Errors:** For calls without an fail assertion broadcast the nature of that error to subscribing features.
    4. **Annotation adjustment:** When working on a fork we can change the annotation from the interface

4.  **System Overview:**

    1. **Architecture information:** An interactive bird's-eye view of the entire system, allowing developers to understand the overall structure and dive deeper into specific modules or verbs.
    2. **Visualizations:**
       1. **Layer Visualization:** Picture each module on its own layer, similar to tracks in a video editing software. On this layered visualization we draw lines or arrows to signify interactions between modules. These connections akin to those in a gantt chart represent what modules have verbs that call other modules.
       2. **Functional flow:** Allows us to understand which modules and verbs get triggered in the by a particular call. This includes highlighting errors resulting from a call.
       3. **Fork:** Indicates to to us that we are working in a branch and not against a production production environment.
    3. **Display filtered/sorted modules:**
       1. **Alphabetic ascending and descending order**
       2. **Historical Evolution:** Displays which modules have had schema changes over the selected period of time.
       3. **Time Range:** Displays only modules deployed and active for a given time scale
    4. **Surface Errors:** The system overview interface uses bold, colors, or flashing icons to make errored verbs stand out prominently. For search param based errors it broadcast the errors to subscribing features.
    5. **Idempotency Tagging:** Verbs can be tagged as idempotent if they are, ensuring that they can be safely retried without side effects.

5.  **Timeline:**

    1. **Display Event Traces:** We are presented a timeline interface of traced **events:**
       1. **Deployments**
       2. **Logs:** We have 5 log levels that are inclusive of the levels below them:
          1. Trace
          2. Debug
          3. Info
          4. Warn
          5. Error
       3. **Calls**
    2. **Display Filtered Traces:** This interface can display to us events based on log levels and other event **parameters**
    3. **Pagination:** We can easily navigate through large amounts of timeline entries via a paginated view and with the ability to adjust the pagination entry count.
    4. **Surface Errors:** The interface uses bold, colors, or flashing icons to make error logs stand out prominently. For search param based errors it broadcast the errors to subscribing features.
    5. **Nested Trace:** Event's originating from another event are nested such as calls
    6. **Tabbed Traces:** Filtered traces can be be saved and tabbed to allow for quickly shifting from a view of all traces to one of the filtered parameters

6.  **Error Management:**

    1. **Subscribes to Error broadcasts:** Subscribes to features that broadcast errors.
    2. **Display Notifications:**
       1. **Improper URL:** When we input a improper URL we will receive a notification.
       2. **Test Call:** When we send an improper test call we will receive a notification.

7.  **Client Side Router:**

    1. **Broadcast routes**: When we make a url change this feature sends messages to other **features** indicating the route changes and search parameters.

8.  **Local Storage Management:**
    1. **Broadcast Storage Updates:** When we make a storage change this feature sends messages to other **features** indicating the local storage has been updated.
    2. **Receive Storage Updates:** This feature can receive message from other features to update values in local storage.
