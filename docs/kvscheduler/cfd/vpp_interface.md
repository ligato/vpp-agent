# Control Flow Diagrams

## Scenario: Create VPP interface via KVDB

This is the most basic scenario covered in the guide. On the other hand,
the attached control-flow diagram is the most verbose - it includes all the
interacting components (listed from the top to the bottom layer):
 * `NB (KVDB)`: contains configuration of a single `my-tap` interface
 * `Orchestrator`: listing KVDB and propagating the configuration to the KVScheduler
 * `KVScheduler`: planning and executing the transaction operations
 * `Interface Descriptor`: implementing CRUD operations for VPP interfaces
 * `Interface Model`: builds and parses keys identifying VPP interfaces


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/add_interface.svg?sanitize=true)