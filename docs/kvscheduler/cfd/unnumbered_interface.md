# Control Flow Diagrams

## Example: Unnumbered interface

Turning interface into unnumbered allows to enable IP processing without assigning
it an explicit IP address. An unnumbered interface can "borrow" the IP address
of another interface, already configured on VPP, which conserves network and
address space.

The requirement is that the interface from which the IP address is supposed
to be borrowed has to be already configured with at least one IP address assigned.
Normally, the agent represents a given VPP interface using a single key-value
pair. Depending on this key alone would only ensure that the target interface is
already configured when a dependent object is being created. in order to be able
to restrict an object existence based on the set of assigned IP addresses to
an interface, every VPP interface value must `derive` single (and unique)
key-value pair for each assigned interface. This will enable to reference IP
address assignments and build dependencies around them.

For unnumbered interfaces alone it would be sufficient to derive single value
from every interface, with key that would allow to determine if the interface
has at least one IP address assigned,
something like: `vpp/interface/<interface-name>/has-ip/<true/false>`
Dependency for an unnumbered interface could then reference key of this value:
`vpp/interface/<interface-name-to-borrow-IP-from>/has-ip/true`

For more complex cases, which are outside of the scope of this example, it may
be desirable to define dependency based not only on the presence but also on the
value of an assigned IP address. Therefore we derive (empty) value for each
assigned IP address with key template:
`"vpp/interface/address/<interface-name>/<address>"`.
This complicates situation for unnumbered interfaces a bit, since they are not
able to reference key of a specific value. Instead, what they would need is to
match any address-representing key, so that the dependency gets satisfied when
at least one of them exists for a given interface. With wildcards this could
be expressed as: `"vpp/interface/address/<interface-name>/*`
The KVScheduler offers even more generic (i.e. expressive) solution than
wildcards: dependency expressed using callback denoted as `AnyOf`. The callback
is a predicate, returning `true` or `false` for a given key. The semantics is
similar to that of the wildcards. The dependency is considered satisfied, when
for at least one of the existing (configured/derived) keys, the callback returns
`true`.

Lastly, to allow a (to-be-)unnumbered interface to exist even if the IP address(es)
to borrow are not available yet, the call to turn interface into unnumbered is
derived from the interface value and processed by a separate descriptor:
`UnnumberedIfDescriptor`. It is this derived value that uses `AnyOf` callback
to trigger IP address borrowing only once the IP addresses become available.
For the time being, the interface is available at least in the L2 mode.

The example also demonstrates that when the borrowed IP address is being removed,
the unnumbered interface will not get un-configured, instead it will only return
the address before it gets unassigned and turn back into the L2 mode.  


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/unnumbered_interface.svg?sanitize=true)