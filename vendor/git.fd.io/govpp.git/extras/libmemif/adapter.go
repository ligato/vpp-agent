// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//	 http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !windows,!darwin

package libmemif

import (
	"encoding/binary"
	"os"
	"sync"
	"syscall"
	"unsafe"

	logger "github.com/Sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lmemif

#include <unistd.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <sys/eventfd.h>
#include <libmemif.h>

// Feature tests.
#ifndef MEMIF_HAVE_CANCEL_POLL_EVENT
// memif_cancel_poll_event that simply returns ErrUnsupported.
static int
memif_cancel_poll_event ()
{
	return 102; // ErrUnsupported
}
#endif

// govpp_memif_conn_args_t replaces fixed sized arrays with C-strings which
// are much easier to work with in cgo.
typedef struct
{
	char *socket_filename;
	char *secret;
	uint8_t num_s2m_rings;
	uint8_t num_m2s_rings;
	uint16_t buffer_size;
	memif_log2_ring_size_t log2_ring_size;
	uint8_t is_master;
	memif_interface_id_t interface_id;
	char *interface_name;
	char *instance_name;
	memif_interface_mode_t mode;
} govpp_memif_conn_args_t;

// govpp_memif_details_t replaces strings represented with (uint8_t *)
// to the standard and easy to work with in cgo: (char *)
typedef struct
{
	char *if_name;
	char *inst_name;
	char *remote_if_name;
	char *remote_inst_name;
	uint32_t id;
	char *secret;
	uint8_t role;
	uint8_t mode;
	char *socket_filename;
	uint8_t rx_queues_num;
	uint8_t tx_queues_num;
	memif_queue_details_t *rx_queues;
	memif_queue_details_t *tx_queues;
	uint8_t link_up_down;
} govpp_memif_details_t;

extern int go_on_connect_callback(void *privateCtx);
extern int go_on_disconnect_callback(void *privateCtx);

// Callbacks strip the connection handle away.

static int
govpp_on_connect_callback(memif_conn_handle_t conn, void *private_ctx)
{
	return go_on_connect_callback(private_ctx);
}

static int
govpp_on_disconnect_callback(memif_conn_handle_t conn, void *private_ctx)
{
	return go_on_disconnect_callback(private_ctx);
}

// govpp_memif_create uses govpp_memif_conn_args_t.
static int
govpp_memif_create (memif_conn_handle_t *conn, govpp_memif_conn_args_t *go_args,
                    void *private_ctx)
{
	memif_conn_args_t args;
	memset (&args, 0, sizeof (args));
	args.socket_filename = (char *)go_args->socket_filename;
	if (go_args->secret != NULL)
	{
		strncpy ((char *)args.secret, go_args->secret,
				 sizeof (args.secret) - 1);
	}
	args.num_s2m_rings = go_args->num_s2m_rings;
	args.num_m2s_rings = go_args->num_m2s_rings;
	args.buffer_size = go_args->buffer_size;
	args.log2_ring_size = go_args->log2_ring_size;
	args.is_master = go_args->is_master;
	args.interface_id = go_args->interface_id;
	if (go_args->interface_name != NULL)
	{
		strncpy ((char *)args.interface_name, go_args->interface_name,
				 sizeof(args.interface_name) - 1);
	}
	if (go_args->instance_name != NULL)
	{
		strncpy ((char *)args.instance_name, go_args->instance_name,
				 sizeof (args.instance_name) - 1);
	}
	args.mode = go_args->mode;

	return memif_create(conn, &args, govpp_on_connect_callback,
						govpp_on_disconnect_callback, NULL,
						private_ctx);
}

// govpp_memif_get_details keeps reallocating buffer until it is large enough.
// The buffer is returned to be deallocated when it is no longer needed.
static int
govpp_memif_get_details (memif_conn_handle_t conn, govpp_memif_details_t *govpp_md,
                         char **buf)
{
	int rv = 0;
	size_t buflen = 1 << 7;
	char *buffer = NULL, *new_buffer = NULL;
	memif_details_t md = {0};

	do {
		// initial malloc (256 bytes) or realloc
		buflen <<= 1;
		new_buffer = realloc(buffer, buflen);
		if (new_buffer == NULL)
		{
			free(buffer);
			return MEMIF_ERR_NOMEM;
		}
		buffer = new_buffer;
		// try to get details
		rv = memif_get_details(conn, &md, buffer, buflen);
	} while (rv == MEMIF_ERR_NOBUF_DET);

	if (rv == 0)
	{
		*buf = buffer;
		govpp_md->if_name = (char *)md.if_name;
		govpp_md->inst_name = (char *)md.inst_name;
		govpp_md->remote_if_name = (char *)md.remote_if_name;
		govpp_md->remote_inst_name = (char *)md.remote_inst_name;
		govpp_md->id = md.id;
		govpp_md->secret = (char *)md.secret;
		govpp_md->role = md.role;
		govpp_md->mode = md.mode;
		govpp_md->socket_filename = (char *)md.socket_filename;
		govpp_md->rx_queues_num = md.rx_queues_num;
		govpp_md->tx_queues_num = md.tx_queues_num;
		govpp_md->rx_queues = md.rx_queues;
		govpp_md->tx_queues = md.tx_queues;
		govpp_md->link_up_down = md.link_up_down;
	}
	else
		free(buffer);
	return rv;
}

// Used to avoid cumbersome tricks that use unsafe.Pointer() + unsafe.Sizeof()
// or even cast C-array directly into Go-slice.
static memif_queue_details_t
govpp_get_rx_queue_details (govpp_memif_details_t *md, int index)
{
	return md->rx_queues[index];
}

// Used to avoid cumbersome tricks that use unsafe.Pointer() + unsafe.Sizeof()
// or even cast C-array directly into Go-slice.
static memif_queue_details_t
govpp_get_tx_queue_details (govpp_memif_details_t *md, int index)
{
	return md->tx_queues[index];
}

// Copy packet data into the selected buffer.
static void
govpp_copy_packet_data(memif_buffer_t *buffers, int index, void *data, uint32_t size)
{
	buffers[index].data_len = (size > buffers[index].buffer_len ? buffers[index].buffer_len : size);
	memcpy(buffers[index].data, data, (size_t)buffers[index].data_len);
}

// Get packet data from the selected buffer.
// Used to avoid an ugly unsafe.Pointer() + unsafe.Sizeof().
static void *
govpp_get_packet_data(memif_buffer_t *buffers, int index, int *size)
{
	*size = (int)buffers[index].data_len;
	return buffers[index].data;
}

*/
import "C"

// IfMode represents the mode (layer/behaviour) in which the interface operates.
type IfMode int

const (
	// IfModeEthernet tells memif to operate on the L2 layer.
	IfModeEthernet IfMode = iota

	// IfModeIP tells memif to operate on the L3 layer.
	IfModeIP

	// IfModePuntInject tells memif to behave as Inject/Punt interface.
	IfModePuntInject
)

// RxMode is used to switch between polling and interrupt for RX.
type RxMode int

const (
	// RxModeInterrupt tells libmemif to send interrupt signal when data are available.
	RxModeInterrupt RxMode = iota

	// RxModePolling means that the user needs to explicitly poll for data on RX
	// queues.
	RxModePolling
)

// RawPacketData represents raw packet data. libmemif doesn't care what the
// actual content is, it only manipulates with raw bytes.
type RawPacketData []byte

// MemifMeta is used to store a basic memif metadata needed for identification
// and connection establishment.
type MemifMeta struct {
	// IfName is the interface name. Has to be unique across all created memifs.
	// Interface name is truncated if needed to have no more than 32 characters.
	IfName string

	// InstanceName identifies the endpoint. If omitted, the application
	// name passed to Init() will be used instead.
	// Instance name is truncated if needed to have no more than 32 characters.
	InstanceName string

	// ConnID is a connection ID used to match opposite sides of the memif
	// connection.
	ConnID uint32

	// SocketFilename is the filename of the AF_UNIX socket through which
	// the connection is established.
	// The string is truncated if neede to fit into sockaddr_un.sun_path
	// (108 characters on Linux).
	SocketFilename string

	// Secret must be the same on both sides for the authentication to succeed.
	// Empty string is allowed.
	// The secret is truncated if needed to have no more than 24 characters.
	Secret string

	// IsMaster is set to true if memif operates in the Master mode.
	IsMaster bool

	// Mode is the mode (layer/behaviour) in which the memif operates.
	Mode IfMode
}

// MemifShmSpecs is used to store the specification of the shared memory segment
// used by memif to send/receive packets.
type MemifShmSpecs struct {
	// NumRxQueues is the number of Rx queues.
	// Default is 1 (used if the value is 0).
	NumRxQueues uint8

	// NumTxQueues is the number of Tx queues.
	// Default is 1 (used if the value is 0).
	NumTxQueues uint8

	// BufferSize is the size of the buffer to hold one packet, or a single
	// fragment of a jumbo frame. Default is 2048 (used if the value is 0).
	BufferSize uint16

	// Log2RingSize is the number of items in the ring represented through
	// the logarithm base 2.
	// Default is 10 (used if the value is 0).
	Log2RingSize uint8
}

// MemifConfig is the memif configuration.
// Used as the input argument to CreateInterface().
// It is the slave's config that mostly decides the parameters of the connection,
// but master may limit some of the quantities if needed (based on the memif
// protocol or master's configuration)
type MemifConfig struct {
	MemifMeta
	MemifShmSpecs
}

// ConnUpdateCallback is a callback type declaration used with callbacks
// related to connection status changes.
type ConnUpdateCallback func(memif *Memif) (err error)

// MemifCallbacks is a container for all callbacks provided by memif.
// Any callback can be nil, in which case it will be simply skipped.
// Important: Do not call CreateInterface() or Memif.Close() from within a callback
// or a deadlock will occur. Instead send signal through a channel to another
// go routine which will be able to create/remove memif interface(s).
type MemifCallbacks struct {
	// OnConnect is triggered when a connection for a given memif was established.
	OnConnect ConnUpdateCallback

	// OnDisconnect is triggered when a connection for a given memif was lost.
	OnDisconnect ConnUpdateCallback
}

// Memif represents a single memif interface. It provides methods to send/receive
// packets in bursts in either the polling mode or in the interrupt mode with
// the help of golang channels.
type Memif struct {
	MemifMeta

	// Per-library references
	ifIndex int                   // index used in the Go-libmemif context (Context.memifs)
	cHandle C.memif_conn_handle_t // handle used in C-libmemif

	// Callbacks
	callbacks *MemifCallbacks

	// Interrupt
	intCh      chan uint8      // memif-global interrupt channel (value = queue ID)
	queueIntCh []chan struct{} // per RX queue interrupt channel

	// Rx/Tx queues
	stopQPollFd int              // event file descriptor used to stop pollRxQueue-s
	wg          sync.WaitGroup   // wait group for all pollRxQueue-s
	rxQueueBufs []CPacketBuffers // an array of C-libmemif packet buffers for each RX queue
	txQueueBufs []CPacketBuffers // an array of C-libmemif packet buffers for each TX queue
}

// MemifDetails provides a detailed runtime information about a memif interface.
type MemifDetails struct {
	MemifMeta
	MemifConnDetails
}

// MemifConnDetails provides a detailed runtime information about a memif
// connection.
type MemifConnDetails struct {
	// RemoteIfName is the name of the memif on the opposite side.
	RemoteIfName string
	// RemoteInstanceName is the name of the endpoint on the opposite side.
	RemoteInstanceName string
	// HasLink is true if the connection has link (= is established and functional).
	HasLink bool
	// RxQueues contains details for each Rx queue.
	RxQueues []MemifQueueDetails
	// TxQueues contains details for each Tx queue.
	TxQueues []MemifQueueDetails
}

// MemifQueueDetails provides a detailed runtime information about a memif queue.
// Queue = Ring + the associated buffers (one directional).
type MemifQueueDetails struct {
	// QueueID is the ID of the queue.
	QueueID uint8
	// RingSize is the number of slots in the ring (not logarithmic).
	RingSize uint32
	// BufferSize is the size of each buffer pointed to from the ring slots.
	BufferSize uint16
	/* Further ring information TO-BE-ADDED when C-libmemif supports them. */
}

// CPacketBuffers stores an array of memif buffers for use with TxBurst or RxBurst.
type CPacketBuffers struct {
	buffers *C.memif_buffer_t
	count   int
}

// Context is a global Go-libmemif runtime context.
type Context struct {
	lock           sync.RWMutex
	initialized    bool
	memifs         map[int] /* ifIndex */ *Memif /* slice of all active memif interfaces */
	nextMemifIndex int

	wg sync.WaitGroup /* wait-group for pollEvents() */
}

var (
	// logger used by the adapter.
	log *logger.Logger

	// Global Go-libmemif context.
	context = &Context{initialized: false}
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	log = logger.New()
	log.Out = os.Stdout
	log.Level = logger.DebugLevel
}

// SetLogger changes the logger for Go-libmemif to the provided one.
// The logger is not used for logging of C-libmemif.
func SetLogger(l *logger.Logger) {
	log = l
}

// Init initializes the libmemif library. Must by called exactly once and before
// any libmemif functions. Do not forget to call Cleanup() before exiting
// your application.
// <appName> should be a human-readable string identifying your application.
// For example, VPP returns the version information ("show version" from VPP CLI).
func Init(appName string) error {
	context.lock.Lock()
	defer context.lock.Unlock()

	if context.initialized {
		return ErrAlreadyInit
	}

	log.Debug("Initializing libmemif library")

	// Initialize C-libmemif.
	var errCode int
	if appName == "" {
		errCode = int(C.memif_init(nil, nil))
	} else {
		appName := C.CString(appName)
		defer C.free(unsafe.Pointer(appName))
		errCode = int(C.memif_init(nil, appName))
	}
	err := getMemifError(errCode)
	if err != nil {
		return err
	}

	// Initialize the map of memory interfaces.
	context.memifs = make(map[int]*Memif)

	// Start event polling.
	context.wg.Add(1)
	go pollEvents()

	context.initialized = true
	log.Debug("libmemif library was initialized")
	return err
}

// Cleanup cleans up all the resources allocated by libmemif.
func Cleanup() error {
	context.lock.Lock()
	defer context.lock.Unlock()

	if !context.initialized {
		return ErrNotInit
	}

	log.Debug("Closing libmemif library")

	// Delete all active interfaces.
	for _, memif := range context.memifs {
		memif.Close()
	}

	// Stop the event loop (if supported by C-libmemif).
	errCode := C.memif_cancel_poll_event()
	err := getMemifError(int(errCode))
	if err == nil {
		log.Debug("Waiting for pollEvents() to stop...")
		context.wg.Wait()
		log.Debug("pollEvents() has stopped...")
	} else {
		log.WithField("err", err).Debug("NOT Waiting for pollEvents to stop...")
	}

	// Run cleanup for C-libmemif.
	err = getMemifError(int(C.memif_cleanup()))
	if err == nil {
		context.initialized = false
		log.Debug("libmemif library was closed")
	}
	return err
}

// CreateInterface creates a new memif interface with the given configuration.
// The same callbacks can be used with multiple memifs. The first callback input
// argument (*Memif) can be used to tell which memif the callback was triggered for.
// The method is thread-safe.
func CreateInterface(config *MemifConfig, callbacks *MemifCallbacks) (memif *Memif, err error) {
	context.lock.Lock()
	defer context.lock.Unlock()

	if !context.initialized {
		return nil, ErrNotInit
	}

	log.WithField("ifName", config.IfName).Debug("Creating a new memif interface")

	// Create memif-wrapper for Go-libmemif.
	memif = &Memif{
		MemifMeta: config.MemifMeta,
		callbacks: &MemifCallbacks{},
		ifIndex:   context.nextMemifIndex,
	}

	// Initialize memif callbacks.
	if callbacks != nil {
		memif.callbacks.OnConnect = callbacks.OnConnect
		memif.callbacks.OnDisconnect = callbacks.OnDisconnect
	}

	// Initialize memif-global interrupt channel.
	memif.intCh = make(chan uint8, 1<<6)

	// Initialize event file descriptor for stopping Rx/Tx queue polling.
	memif.stopQPollFd = int(C.eventfd(0, C.EFD_NONBLOCK))
	if memif.stopQPollFd < 0 {
		return nil, ErrSyscall
	}

	// Initialize memif input arguments.
	args := &C.govpp_memif_conn_args_t{}
	// - socket file name
	if config.SocketFilename != "" {
		args.socket_filename = C.CString(config.SocketFilename)
		defer C.free(unsafe.Pointer(args.socket_filename))
	}
	// - interface ID
	args.interface_id = C.memif_interface_id_t(config.ConnID)
	// - interface name
	if config.IfName != "" {
		args.interface_name = C.CString(config.IfName)
		defer C.free(unsafe.Pointer(args.interface_name))
	}
	// - instance name
	if config.InstanceName != "" {
		args.instance_name = C.CString(config.InstanceName)
		defer C.free(unsafe.Pointer(args.instance_name))
	}
	// - mode
	switch config.Mode {
	case IfModeEthernet:
		args.mode = C.MEMIF_INTERFACE_MODE_ETHERNET
	case IfModeIP:
		args.mode = C.MEMIF_INTERFACE_MODE_IP
	case IfModePuntInject:
		args.mode = C.MEMIF_INTERFACE_MODE_PUNT_INJECT
	default:
		args.mode = C.MEMIF_INTERFACE_MODE_ETHERNET
	}
	// - secret
	if config.Secret != "" {
		args.secret = C.CString(config.Secret)
		defer C.free(unsafe.Pointer(args.secret))
	}
	// - master/slave flag + number of Rx/Tx queues
	if config.IsMaster {
		args.num_s2m_rings = C.uint8_t(config.NumRxQueues)
		args.num_m2s_rings = C.uint8_t(config.NumTxQueues)
		args.is_master = C.uint8_t(1)
	} else {
		args.num_s2m_rings = C.uint8_t(config.NumTxQueues)
		args.num_m2s_rings = C.uint8_t(config.NumRxQueues)
		args.is_master = C.uint8_t(0)
	}
	// - buffer size
	args.buffer_size = C.uint16_t(config.BufferSize)
	// - log_2(ring size)
	args.log2_ring_size = C.memif_log2_ring_size_t(config.Log2RingSize)

	// Create memif in C-libmemif.
	errCode := C.govpp_memif_create(&memif.cHandle, args, unsafe.Pointer(uintptr(memif.ifIndex)))
	err = getMemifError(int(errCode))
	if err != nil {
		return nil, err
	}

	// Register the new memif.
	context.memifs[memif.ifIndex] = memif
	context.nextMemifIndex++
	log.WithField("ifName", config.IfName).Debug("A new memif interface was created")

	return memif, nil
}

// GetInterruptChan returns a channel which is continuously being filled with
// IDs of queues with data ready to be received.
// Since there is only one interrupt signal sent for an entire burst of packets,
// an interrupt handling routine should repeatedly call RxBurst() until
// the function returns an empty slice of packets. This way it is ensured
// that there are no packets left on the queue unread when the interrupt signal
// is cleared.
// The method is thread-safe.
func (memif *Memif) GetInterruptChan() (ch <-chan uint8 /* queue ID */) {
	return memif.intCh
}

// GetQueueInterruptChan returns an empty-data channel which fires every time
// there are data to read on a given queue.
// It is only valid to call this function if memif is in the connected state.
// Channel is automatically closed when the connection goes down (but after
// the user provided callback OnDisconnect has executed).
// Since there is only one interrupt signal sent for an entire burst of packets,
// an interrupt handling routine should repeatedly call RxBurst() until
// the function returns an empty slice of packets. This way it is ensured
// that there are no packets left on the queue unread when the interrupt signal
// is cleared.
// The method is thread-safe.
func (memif *Memif) GetQueueInterruptChan(queueID uint8) (ch <-chan struct{}, err error) {
	if int(queueID) >= len(memif.queueIntCh) {
		return nil, ErrQueueID
	}
	return memif.queueIntCh[queueID], nil
}

// SetRxMode allows to switch between the interrupt and the polling mode for Rx.
// The method is thread-safe.
func (memif *Memif) SetRxMode(queueID uint8, rxMode RxMode) (err error) {
	var cRxMode C.memif_rx_mode_t
	switch rxMode {
	case RxModeInterrupt:
		cRxMode = C.MEMIF_RX_MODE_INTERRUPT
	case RxModePolling:
		cRxMode = C.MEMIF_RX_MODE_POLLING
	default:
		cRxMode = C.MEMIF_RX_MODE_INTERRUPT
	}
	errCode := C.memif_set_rx_mode(memif.cHandle, cRxMode, C.uint16_t(queueID))
	return getMemifError(int(errCode))
}

// GetDetails returns a detailed runtime information about this memif.
// The method is thread-safe.
func (memif *Memif) GetDetails() (details *MemifDetails, err error) {
	cDetails := C.govpp_memif_details_t{}
	var buf *C.char

	// Get memif details from C-libmemif.
	errCode := C.govpp_memif_get_details(memif.cHandle, &cDetails, &buf)
	err = getMemifError(int(errCode))
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(buf))

	// Convert details from C to Go.
	details = &MemifDetails{}
	// - metadata:
	details.IfName = C.GoString(cDetails.if_name)
	details.InstanceName = C.GoString(cDetails.inst_name)
	details.ConnID = uint32(cDetails.id)
	details.SocketFilename = C.GoString(cDetails.socket_filename)
	if cDetails.secret != nil {
		details.Secret = C.GoString(cDetails.secret)
	}
	details.IsMaster = cDetails.role == C.uint8_t(0)
	switch cDetails.mode {
	case C.MEMIF_INTERFACE_MODE_ETHERNET:
		details.Mode = IfModeEthernet
	case C.MEMIF_INTERFACE_MODE_IP:
		details.Mode = IfModeIP
	case C.MEMIF_INTERFACE_MODE_PUNT_INJECT:
		details.Mode = IfModePuntInject
	default:
		details.Mode = IfModeEthernet
	}
	// - connection details:
	details.RemoteIfName = C.GoString(cDetails.remote_if_name)
	details.RemoteInstanceName = C.GoString(cDetails.remote_inst_name)
	details.HasLink = cDetails.link_up_down == C.uint8_t(1)
	// - RX queues:
	var i uint8
	for i = 0; i < uint8(cDetails.rx_queues_num); i++ {
		cRxQueue := C.govpp_get_rx_queue_details(&cDetails, C.int(i))
		queueDetails := MemifQueueDetails{
			QueueID:    uint8(cRxQueue.qid),
			RingSize:   uint32(cRxQueue.ring_size),
			BufferSize: uint16(cRxQueue.buffer_size),
		}
		details.RxQueues = append(details.RxQueues, queueDetails)
	}
	// - TX queues:
	for i = 0; i < uint8(cDetails.tx_queues_num); i++ {
		cTxQueue := C.govpp_get_tx_queue_details(&cDetails, C.int(i))
		queueDetails := MemifQueueDetails{
			QueueID:    uint8(cTxQueue.qid),
			RingSize:   uint32(cTxQueue.ring_size),
			BufferSize: uint16(cTxQueue.buffer_size),
		}
		details.TxQueues = append(details.TxQueues, queueDetails)
	}

	return details, nil
}

// TxBurst is used to send multiple packets in one call into a selected queue.
// The actual number of packets sent may be smaller and is returned as <count>.
// The method is non-blocking even if the ring is full and no packet can be sent.
// It is only valid to call this function if memif is in the connected state.
// Multiple TxBurst-s can run concurrently provided that each targets a different
// TX queue.
func (memif *Memif) TxBurst(queueID uint8, packets []RawPacketData) (count uint16, err error) {
	var sentCount C.uint16_t
	var allocated C.uint16_t
	var bufSize int

	if len(packets) == 0 {
		return 0, nil
	}

	if int(queueID) >= len(memif.txQueueBufs) {
		return 0, ErrQueueID
	}

	// The largest packet in the set determines the packet buffer size.
	for _, packet := range packets {
		if len(packet) > int(bufSize) {
			bufSize = len(packet)
		}
	}

	// Reallocate Tx buffers if needed to fit the input packets.
	pb := memif.txQueueBufs[queueID]
	bufCount := len(packets)
	if pb.count < bufCount {
		newBuffers := C.realloc(unsafe.Pointer(pb.buffers), C.size_t(bufCount*int(C.sizeof_memif_buffer_t)))
		if newBuffers == nil {
			// Realloc failed, <count> will be less than len(packets).
			bufCount = pb.count
		} else {
			pb.buffers = (*C.memif_buffer_t)(newBuffers)
			pb.count = bufCount
		}
	}

	// Allocate ring slots.
	cQueueID := C.uint16_t(queueID)
	errCode := C.memif_buffer_alloc(memif.cHandle, cQueueID, pb.buffers, C.uint16_t(bufCount),
		&allocated, C.uint16_t(bufSize))
	err = getMemifError(int(errCode))
	if err == ErrNoBufRing {
		// Not enough ring slots, <count> will be less than bufCount.
		err = nil
	}
	if err != nil {
		return 0, err
	}

	// Copy packet data into the buffers.
	for i := 0; i < int(allocated); i++ {
		packetData := unsafe.Pointer(&packets[i][0])
		C.govpp_copy_packet_data(pb.buffers, C.int(i), packetData, C.uint32_t(len(packets[i])))
	}

	errCode = C.memif_tx_burst(memif.cHandle, cQueueID, pb.buffers, allocated, &sentCount)
	err = getMemifError(int(errCode))
	if err != nil {
		return 0, err
	}
	count = uint16(sentCount)

	return count, nil
}

// RxBurst is used to receive multiple packets in one call from a selected queue.
// <count> is the number of packets to receive. The actual number of packets
// received may be smaller. <count> effectively limits the maximum number
// of packets to receive in one burst (for a flat, predictable memory usage).
// The method is non-blocking even if there are no packets to receive.
// It is only valid to call this function if memif is in the connected state.
// Multiple RxBurst-s can run concurrently provided that each targets a different
// Rx queue.
func (memif *Memif) RxBurst(queueID uint8, count uint16) (packets []RawPacketData, err error) {
	var recvCount C.uint16_t
	var freed C.uint16_t

	if count == 0 {
		return packets, nil
	}

	if int(queueID) >= len(memif.rxQueueBufs) {
		return packets, ErrQueueID
	}

	// Reallocate Rx buffers if needed to fit the output packets.
	pb := memif.rxQueueBufs[queueID]
	bufCount := int(count)
	if pb.count < bufCount {
		newBuffers := C.realloc(unsafe.Pointer(pb.buffers), C.size_t(bufCount*int(C.sizeof_memif_buffer_t)))
		if newBuffers == nil {
			// Realloc failed, len(<packets>) will be certainly less than <count>.
			bufCount = pb.count
		} else {
			pb.buffers = (*C.memif_buffer_t)(newBuffers)
			pb.count = bufCount
		}
	}

	cQueueID := C.uint16_t(queueID)
	errCode := C.memif_rx_burst(memif.cHandle, cQueueID, pb.buffers, C.uint16_t(bufCount), &recvCount)
	err = getMemifError(int(errCode))
	if err == ErrNoBuf {
		// More packets to read - the user is expected to run RxBurst() until there
		// are no more packets to receive.
		err = nil
	}
	if err != nil {
		return packets, err
	}

	// Copy packet data into the instances of RawPacketData.
	for i := 0; i < int(recvCount); i++ {
		var packetSize C.int
		packetData := C.govpp_get_packet_data(pb.buffers, C.int(i), &packetSize)
		packets = append(packets, C.GoBytes(packetData, packetSize))
	}

	errCode = C.memif_buffer_free(memif.cHandle, cQueueID, pb.buffers, recvCount, &freed)
	err = getMemifError(int(errCode))
	if err != nil {
		// Throw away packets to avoid duplicities.
		packets = nil
	}

	return packets, err
}

// Close removes the memif interface. If the memif is in the connected state,
// the connection is first properly closed.
// Do not access memif after it is closed, let garbage collector to remove it.
func (memif *Memif) Close() error {
	log.WithField("ifName", memif.IfName).Debug("Closing the memif interface")

	// Delete memif from C-libmemif.
	err := getMemifError(int(C.memif_delete(&memif.cHandle)))

	if err != nil {
		// Close memif-global interrupt channel.
		close(memif.intCh)
		// Close file descriptor stopQPollFd.
		C.close(C.int(memif.stopQPollFd))
	}

	context.lock.Lock()
	defer context.lock.Unlock()
	// Unregister the interface from the context.
	delete(context.memifs, memif.ifIndex)
	log.WithField("ifName", memif.IfName).Debug("memif interface was closed")

	return err
}

// initQueues allocates resources associated with Rx/Tx queues.
func (memif *Memif) initQueues() error {
	// Get Rx/Tx queues count.
	details, err := memif.GetDetails()
	if err != nil {
		return err
	}

	log.WithFields(logger.Fields{
		"ifName":   memif.IfName,
		"Rx-count": len(details.RxQueues),
		"Tx-count": len(details.TxQueues),
	}).Debug("Initializing Rx/Tx queues.")

	// Initialize interrupt channels.
	var i int
	for i = 0; i < len(details.RxQueues); i++ {
		queueIntCh := make(chan struct{}, 1)
		memif.queueIntCh = append(memif.queueIntCh, queueIntCh)
	}

	// Initialize Rx/Tx packet buffers.
	for i = 0; i < len(details.RxQueues); i++ {
		memif.rxQueueBufs = append(memif.rxQueueBufs, CPacketBuffers{})
	}
	for i = 0; i < len(details.TxQueues); i++ {
		memif.txQueueBufs = append(memif.txQueueBufs, CPacketBuffers{})
	}

	return nil
}

// closeQueues deallocates all resources associated with Rx/Tx queues.
func (memif *Memif) closeQueues() {
	log.WithFields(logger.Fields{
		"ifName":   memif.IfName,
		"Rx-count": len(memif.rxQueueBufs),
		"Tx-count": len(memif.txQueueBufs),
	}).Debug("Closing Rx/Tx queues.")

	// Close interrupt channels.
	for _, ch := range memif.queueIntCh {
		close(ch)
	}
	memif.queueIntCh = nil

	// Deallocate Rx/Tx packet buffers.
	for _, pb := range memif.rxQueueBufs {
		C.free(unsafe.Pointer(pb.buffers))
	}
	memif.rxQueueBufs = nil
	for _, pb := range memif.txQueueBufs {
		C.free(unsafe.Pointer(pb.buffers))
	}
	memif.txQueueBufs = nil
}

// pollEvents repeatedly polls for a libmemif event.
func pollEvents() {
	defer context.wg.Done()
	for {
		errCode := C.memif_poll_event(C.int(-1))
		err := getMemifError(int(errCode))
		if err == ErrPollCanceled {
			return
		}
	}
}

// pollRxQueue repeatedly polls an Rx queue for interrupts.
func pollRxQueue(memif *Memif, queueID uint8) {
	defer memif.wg.Done()

	log.WithFields(logger.Fields{
		"ifName":   memif.IfName,
		"queue-ID": queueID,
	}).Debug("Started queue interrupt polling.")

	var qfd C.int
	errCode := C.memif_get_queue_efd(memif.cHandle, C.uint16_t(queueID), &qfd)
	err := getMemifError(int(errCode))
	if err != nil {
		log.WithField("err", err).Error("memif_get_queue_efd() failed")
		return
	}

	// Create epoll file descriptor.
	var event [1]syscall.EpollEvent
	epFd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.WithField("err", err).Error("epoll_create1() failed")
		return
	}
	defer syscall.Close(epFd)

	// Add Rx queue interrupt file descriptor.
	event[0].Events = syscall.EPOLLIN
	event[0].Fd = int32(qfd)
	if err = syscall.EpollCtl(epFd, syscall.EPOLL_CTL_ADD, int(qfd), &event[0]); err != nil {
		log.WithField("err", err).Error("epoll_ctl() failed")
		return
	}

	// Add file descriptor used to stop this go routine.
	event[0].Events = syscall.EPOLLIN
	event[0].Fd = int32(memif.stopQPollFd)
	if err = syscall.EpollCtl(epFd, syscall.EPOLL_CTL_ADD, memif.stopQPollFd, &event[0]); err != nil {
		log.WithField("err", err).Error("epoll_ctl() failed")
		return
	}

	// Poll for interrupts.
	for {
		_, err := syscall.EpollWait(epFd, event[:], -1)
		if err != nil {
			log.WithField("err", err).Error("epoll_wait() failed")
			return
		}

		// Handle Rx Interrupt.
		if event[0].Fd == int32(qfd) {
			// Consume the interrupt event.
			buf := make([]byte, 8)
			_, err = syscall.Read(int(qfd), buf[:])
			if err != nil {
				log.WithField("err", err).Warn("read() failed")
			}

			// Send signal to memif-global interrupt channel.
			select {
			case memif.intCh <- queueID:
				break
			default:
				break
			}

			// Send signal to queue-specific interrupt channel.
			select {
			case memif.queueIntCh[queueID] <- struct{}{}:
				break
			default:
				break
			}
		}

		// Stop the go routine if requested.
		if event[0].Fd == int32(memif.stopQPollFd) {
			log.WithFields(logger.Fields{
				"ifName":   memif.IfName,
				"queue-ID": queueID,
			}).Debug("Stopped queue interrupt polling.")
			return
		}
	}
}

//export go_on_connect_callback
func go_on_connect_callback(privateCtx unsafe.Pointer) C.int {
	log.Debug("go_on_connect_callback BEGIN")
	defer log.Debug("go_on_connect_callback END")
	context.lock.RLock()
	defer context.lock.RUnlock()

	// Get memif reference.
	ifIndex := int(uintptr(privateCtx))
	memif, exists := context.memifs[ifIndex]
	if !exists {
		return C.int(ErrNoConn.Code())
	}

	// Initialize Rx/Tx queues.
	err := memif.initQueues()
	if err != nil {
		if memifErr, ok := err.(*MemifError); ok {
			return C.int(memifErr.Code())
		}
		return C.int(ErrUnknown.Code())
	}

	// Call the user callback.
	if memif.callbacks.OnConnect != nil {
		memif.callbacks.OnConnect(memif)
	}

	// Start polling the RX queues for interrupts.
	for i := 0; i < len(memif.queueIntCh); i++ {
		memif.wg.Add(1)
		go pollRxQueue(memif, uint8(i))
	}

	return C.int(0)
}

//export go_on_disconnect_callback
func go_on_disconnect_callback(privateCtx unsafe.Pointer) C.int {
	log.Debug("go_on_disconnect_callback BEGIN")
	defer log.Debug("go_on_disconnect_callback END")
	context.lock.RLock()
	defer context.lock.RUnlock()

	// Get memif reference.
	ifIndex := int(uintptr(privateCtx))
	memif, exists := context.memifs[ifIndex]
	if !exists {
		// Already closed.
		return C.int(0)
	}

	// Stop polling the RX queues for interrupts.
	buf := make([]byte, 8)
	binary.PutUvarint(buf, 1)
	// - add an event
	_, err := syscall.Write(memif.stopQPollFd, buf[:])
	if err != nil {
		return C.int(ErrSyscall.Code())
	}
	// - wait
	memif.wg.Wait()
	// - remove the event
	_, err = syscall.Read(memif.stopQPollFd, buf[:])
	if err != nil {
		return C.int(ErrSyscall.Code())
	}

	// Call the user callback.
	if memif.callbacks.OnDisconnect != nil {
		memif.callbacks.OnDisconnect(memif)
	}

	// Close Rx/Tx queues.
	memif.closeQueues()

	return C.int(0)
}
