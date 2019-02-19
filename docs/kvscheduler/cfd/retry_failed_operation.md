# Control Flow Diagrams

## Example: Retry of failed operation

This example demonstrates that for `best-effort` transactions (e.g. every resync),
the KVScheduler allows to enable automatic and potentially repeated *retry*
of failed operations.

In this case, a TAP interface `my-tap` fails to get created. Before terminating
the transaction, the scheduler retrieves the current value of `my-value`, the
state of which cannot be assumed since the creation failed somewhere in-progress.
Then it schedules a so-called *retry transaction*, which will attempt to fix
the failure by re-applying the same configuration. Since the value retrieval
has not found `my-tap` to be configured, the retry transaction will repeat
the `Create(my-tap)` operation and succeed in our example.


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/fails_to_add_interface.svg?sanitize=true)