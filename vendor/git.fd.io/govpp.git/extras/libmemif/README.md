## Go-libmemif

Package **libmemif** is a Golang adapter for the **libmemif library**
(`extras/libmemif` in the [VPP](https://wiki.fd.io/view/VPP) repository).
To differentiate between the adapter and the underlying C-written library,
labels `Go-libmemif` and `C-libmemif` are used in the documentation.

### Requirements

libmemif for Golang is build on the top of the original, C-written
libmemif library using `cgo`. It is therefore necessary to have C-libmemif
header files and the library itself installed in locations known
to the compiler.

For example, to install C-libmemif system-wide into the standard
locations, execute:
```
$ git clone https://gerrit.fd.io/r/vpp
$ cd vpp/extras/libmemif
$ make install
```

### Build

Package **libmemif** is not part of the **GoVPP** core and as such it is
not included in the [make build](../../Makefile) target.
Instead, it has its own target in the [top-level Makefile](../../Makefile)
used to build the attached examples with the adapter:
```
$ make extras
```

### APIs

All **Go-libmemif** public APIs can be found in [adapter.go](adapter.go).
Please see the comments for a more detailed description.
Additionally, a list of all errors thrown by libmemif can be found
in [error.go](error.go).

### Usage

**libmemif** needs to be first initialized with `Init(appName)`.
This has to be done only once in the context of the entire process.
Make sure to call `Cleanup()` to release all the resources allocated
by **libmemif** before exiting your application. Consider calling
`Init()` followed by `Cleanup()` scheduled with `defer` in the `main()`
function.

Log messages are by default printed to stdout. Use `SetLogger()` to use
your own customized logger (can be changed before `Init()`).

Once **libmemif** is initialized, new memif interfaces can be created
with `CreateInterface(config, callbacks)`. See `MemifConfig` structure
definition to learn about possible memif configuration options.
If successful, `CreateInterface()` returns an instance of `Memif`
structure representing the underlying memif interface.

Callbacks are optional and can be shared across multiple memif instances.
Available callbacks are:
1. **OnConnect**: called when the connection is established.
   By the time the callback is called, the Rx/Tx queues are initialized
   and ready for data transmission. Interrupt channels are also
   created and ready to be read from.
   The user is expected to start polling for input packets via repeated
   calls to `Memif.RxBurst(queueID, count)` or to initiate select
   on the interrupt channels obtained with `Get*InterruptChan()`,
   depending on the Rx mode. By default, all memif Rx queues are created
   in the interrupt mode, but this can be changed per-queue with
   `Memif.SetRxMode(queueID, mode)`.
2. **OnDisconnect**: called after the connection was closed. Immediately
   after the user callback returns, Rx/Tx queues and interrupt channels
   are also deallocated. The user defined callback should therefore ensure
   that all the Rx/Tx operations are stopped before it returns.

**libmemif** was designed for a maximum possible performance. Packets
are sent and received in bulks, rather than one-by-one, using
`Memif.TxBurst(queueID, packets)` and `Memif.RxBurst(queueID, count)`,
respectively. Memif connection can consists of multiple queues in both
directions. A queue is one-directional wait-free ring buffer.
It is the unit of parallelism for data transmission. The maximum possible
lock-free granularity is therefore one go routine for one queue.

Interrupt channel for one specific Rx queue can be obtained with
`GetQueueInterruptChan(queueID)` as opposed to `GetInterruptChan()`
for all the Rx queues. There is only one interrupt signal sent for
an entire burst of packets, therefore an interrupt handling routine
should repeatedly call RxBurst() until an empty slice of packets
is returned. This way it is ensured that there are no packets left
on the queue unread when the interrupt signal is cleared.
Study the `ReadAndPrintPackets()` function in [raw-data example](examples/raw-data/raw-data.go).

For **libmemif** the packet is just an array of bytes. It does not care
what the actual content is. It is not required for a packet to follow
any network protocol in order to get transported from one end to another.
See the type declaration for `RawPacketData` and its use in `Memif.TxBurst()`
and `Memif.RxBurst()`.

In order to remove a memif interface, call `Memif.Close()`. If the memif
is in the connected state, the connection is first properly closed.
Do not touch memif after it was closed, let garbage collector to remove
the `Memif` instance. In the end, `Cleanup()` will also ensure that all
active memif interfaces are closed before the cleanup finalizes.

### Examples

**Go-libmemif** ships with two simple examples demonstrating the usage
of the package with a detailed commentary.
The examples can be found in the subdirectory [examples](./examples).

#### Raw data (libmemif <-> libmemif)

*raw-data* is a basic example showing how to create a memif interface,
handle events through callbacks and perform Rx/Tx of raw data. Before
handling an actual packets it is important to understand the skeleton
of libmemif-based applications.

Since VPP expects proper packet data, it is not very useful to connect
*raw-data* example with VPP, even though it will work, since all
the received data will get dropped on the VPP side.

To create a connection of two raw-data instances, start two processes
concurrently in an arbitrary order:
 - *master* memif:
   ```
   $ cd extras/libmemif/examples/raw-data
   $ ./raw-data
   ```
 - *slave* memif:
   ```
   $ cd extras/libmemif/examples/raw-data
   $ ./raw-data --slave
   ```

Every 3 seconds both sides send 3 raw-data packets to the opposite end
through each of the 3 queues. The received packets are printed to stdout.

Stop an instance of *raw-data* with an interrupt signal (^C).

#### ICMP Responder

*icmp-responder* is a simple example showing how to answer APR and ICMP
echo requests through a memif interface. Package `google/gopacket` is
used to decode and construct packets.

The appropriate VPP configuration for the opposite memif is:
```
vpp$ create memif id 1 socket /tmp/icmp-responder-example slave secret secret
vpp$ set int state memif0/1 up
vpp$ set int ip address memif0/1 192.168.1.2/24
```

To start the example, simply type:
```
root$ ./icmp-responder
```

*icmp-responder* needs to be run as root so that it can access the socket
created by VPP.

Normally, the memif interface is in the master mode. Pass CLI flag `--slave`
to create memif in the slave mode:
```
root$ ./icmp-responder --slave
```

Don't forget to put the opposite memif into the master mode in that case.

To verify the connection, run:
```
vpp$ ping 192.168.1.1
64 bytes from 192.168.1.1: icmp_seq=2 ttl=255 time=.6974 ms
64 bytes from 192.168.1.1: icmp_seq=3 ttl=255 time=.6310 ms
64 bytes from 192.168.1.1: icmp_seq=4 ttl=255 time=1.0350 ms
64 bytes from 192.168.1.1: icmp_seq=5 ttl=255 time=.5359 ms

Statistics: 5 sent, 4 received, 20% packet loss
vpp$ sh ip arp
    Time           IP4       Flags      Ethernet              Interface
    68.5648   192.168.1.1     D    aa:aa:aa:aa:aa:aa memif0/1
```
*Note*: it is expected that the first ping is shown as lost.
        It was actually converted to an ARP request. This is a VPP
        specific feature common to all interface types.

Stop the example with an interrupt signal (^C).