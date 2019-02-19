# Mock Plugins

This is an interactive, hands-on demonstration of the KVScheduler framework,
based on replicated `vpp/ifplugin` and `vpp/l2plugin`, where models were
significantly simplified and instead of the actual VPP, a mock southbound plane
was implemented, printing the triggered CRUD operations into the stdout
instead of actually executing them.
This allows you to run the example as a standalone process, without any
prerequisites.

The focus is on the KVScheduler and how it operates and interacts with the
registered descriptors under various scenarios. From that point of view,
the SB is irrelevant - the framework abstracts from the downstream interactions.

Furthermore, by getting rid of all the VPP-specific implementation details,
the plugins became minimalistic and can serve as templates for the development
of new plugins for VPP, Linux or even for SB completely new to the agent.
The code of replicated plugins is well-documented and should help you to
understand why certain key aspects of plugins/descriptors/models are implemented
the way they are.

Build and run the example with:
```
go build .
./mock_plugins
```
You will be asked to select a scenario from a provided set to execute.