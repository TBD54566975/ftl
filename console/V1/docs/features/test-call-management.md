# Test Call Management

- Use Case 1
- Context:
  - verb code development
  - verb code debugging

## Requirements

1. **Select a Verb:**

   - Display a list of available verbs from the selected module.
   - Allow the user to select a specific verb for which they want to send a test call.

2. **Input Test Call Parameters:**
   - Provide an intuitive interface for the user to input or select the necessary parameters for the test call. This might include headers, payloads, etc.
   - Display the expected schema for the selected verb to guide the user.
3. **Send Test Call:**

   - Have a "Send" or "Execute" button to initiate the test call.
   - Display real-time feedback, possibly in a console or log format, to show the progress and response of the test call.

4. **Error Handling:**
   - In case of errors, display a detailed error message.
   - Offer suggestions or possible fixes for common errors.
   - Include an option to quickly resend the test call with the observed signature, as mentioned in the brief.
5. **Save Test Call:**

   - Allow users to save the test call configuration (including parameters and response) for future reference.
   - Assign a unique name or identifier for each saved test call for easy retrieval.

6. **Manage Saved Test Calls:**

   - Display a list or library of saved test calls.
   - Provide options to:
     - View details of a saved test call.
     - Edit or modify a saved test call.
     - Delete a saved test call.
     - Clone a saved test call to create a similar new one.

7. **Switch Between Test Calls:**

   - Allow users to quickly switch between different saved test calls.
   - Provide a dropdown or list for easy navigation between them.

8. **Resend Test Calls:**
   - Provide an option to quickly resend any saved test call.
   - Display the response and any differences from the previous execution.
