# FTL fsm example

Run using:
```sh
> call fsm.sendEvent '{"id":"door3", "event":"open"}'
debug:fsm:runner5: The door is open.

> call fsm.sendEvent '{"id":"door3", "event":"jam"}'
debug:fsm:runner5: The door is jammed. Fixing...

> call fsm.sendEvent '{"id":"door3", "event":"unlock"}'
debug:fsm:runner5: The door is unlocked.

> call fsm.sendEvent '{"id":"door3", "event":"lock"}'
debug:fsm:runner5: The door is locked.

> call fsm.sendEvent '{"id":"door3", "event":"jam"}'
debug:fsm:runner5: The door is jammed. Fixing...

> call fsm.sendEvent '{"id":"door3", "event":"lock"}'
error:fsm:runner5: Call to deployments dpl-fsm-7vye13artgx3hkx failed: call to verb fsm.sendEvent failed: failed to send fsm event: failed_precondition: no transition found from state fsm.openDoor for type fsm.Locked, candidates are unlockDoor, jamDoor
```