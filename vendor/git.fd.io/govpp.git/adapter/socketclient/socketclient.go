package socketclient

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lunixbochs/struc"
	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/codec"
	"git.fd.io/govpp.git/examples/bin_api/memclnt"
)

const (
	// DefaultSocketName is default VPP API socket file name
	DefaultSocketName = "/run/vpp-api.sock"

	sockCreateMsgId = 15          // hard-coded id for sockclnt_create message
	govppClientName = "govppsock" // client name used for socket registration
)

var (
	ConnectTimeout    = time.Second * 3
	DisconnectTimeout = time.Second

	Debug       = os.Getenv("DEBUG_GOVPP_SOCK") != ""
	DebugMsgIds = os.Getenv("DEBUG_GOVPP_SOCKMSG") != ""

	Log = logger.New() // global logger
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	Log.Out = os.Stdout
	if Debug {
		Log.Level = logger.DebugLevel
	}
}

type vppClient struct {
	sockAddr     string
	conn         *net.UnixConn
	reader       *bufio.Reader
	cb           adapter.MsgCallback
	clientIndex  uint32
	msgTable     map[string]uint16
	sockDelMsgId uint16
	writeMu      sync.Mutex
	quit         chan struct{}
	wg           sync.WaitGroup
}

func NewVppClient(sockAddr string) *vppClient {
	if sockAddr == "" {
		sockAddr = DefaultSocketName
	}
	return &vppClient{
		sockAddr: sockAddr,
		cb:       nilCallback,
	}
}

func nilCallback(msgID uint16, data []byte) {
	Log.Warnf("no callback set, dropping message: ID=%v len=%d", msgID, len(data))
}

// WaitReady checks socket file existence and waits for it if necessary
func (c *vppClient) WaitReady() error {
	// verify file existence
	if _, err := os.Stat(c.sockAddr); err == nil {
		return nil
	} else if os.IsExist(err) {
		return err
	}

	// if not, watch for it
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			Log.Errorf("failed to close file watcher: %v", err)
		}
	}()
	path := filepath.Dir(c.sockAddr)
	if err := watcher.Add(path); err != nil {
		return err
	}

	for {
		ev := <-watcher.Events
		if ev.Name == path {
			if (ev.Op & fsnotify.Create) == fsnotify.Create {
				// socket ready
				return nil
			}
		}
	}
}

func (c *vppClient) SetMsgCallback(cb adapter.MsgCallback) {
	Log.Debug("SetMsgCallback")
	c.cb = cb
}

func (c *vppClient) Connect() error {
	Log.Debugf("Connecting to: %v", c.sockAddr)

	if err := c.connect(c.sockAddr); err != nil {
		return err
	}

	if err := c.open(); err != nil {
		return err
	}

	c.quit = make(chan struct{})
	c.wg.Add(1)
	go c.readerLoop()

	return nil
}

func (c *vppClient) connect(sockAddr string) error {
	addr, err := net.ResolveUnixAddr("unixpacket", sockAddr)
	if err != nil {
		Log.Debugln("ResolveUnixAddr error:", err)
		return err
	}

	conn, err := net.DialUnix("unixpacket", nil, addr)
	if err != nil {
		Log.Debugln("Dial error:", err)
		return err
	}

	c.conn = conn
	c.reader = bufio.NewReader(c.conn)

	Log.Debugf("Connected to socket: %v", addr)

	return nil
}

func (c *vppClient) open() error {
	msgCodec := new(codec.MsgCodec)

	req := &memclnt.SockclntCreate{
		Name: []byte(govppClientName),
	}
	msg, err := msgCodec.EncodeMsg(req, sockCreateMsgId)
	if err != nil {
		Log.Debugln("Encode error:", err)
		return err
	}
	// set non-0 context
	msg[5] = 123

	if err := c.write(msg); err != nil {
		Log.Debugln("Write error: ", err)
		return err
	}

	readDeadline := time.Now().Add(ConnectTimeout)
	if err := c.conn.SetReadDeadline(readDeadline); err != nil {
		return err
	}
	msgReply, err := c.read()
	if err != nil {
		Log.Println("Read error:", err)
		return err
	}
	// reset read deadline
	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	//log.Printf("Client got (%d): % 0X", len(msgReply), msgReply)

	reply := new(memclnt.SockclntCreateReply)
	if err := msgCodec.DecodeMsg(msgReply, reply); err != nil {
		Log.Println("Decode error:", err)
		return err
	}

	Log.Debugf("SockclntCreateReply: Response=%v Index=%v Count=%v",
		reply.Response, reply.Index, reply.Count)

	c.clientIndex = reply.Index
	c.msgTable = make(map[string]uint16, reply.Count)
	for _, x := range reply.MessageTable {
		name := string(bytes.TrimSuffix(bytes.Split(x.Name, []byte{0x00})[0], []byte{0x13}))
		c.msgTable[name] = x.Index
		if strings.HasPrefix(name, "sockclnt_delete_") {
			c.sockDelMsgId = x.Index
		}
		if DebugMsgIds {
			Log.Debugf(" - %4d: %q", x.Index, name)
		}
	}

	return nil
}

func (c *vppClient) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	Log.Debugf("Disconnecting..")

	close(c.quit)

	// force readerLoop to timeout
	if err := c.conn.SetReadDeadline(time.Now()); err != nil {
		return err
	}

	// wait for readerLoop to return
	c.wg.Wait()

	if err := c.close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		Log.Debugln("Close socket conn failed:", err)
		return err
	}

	return nil
}

func (c *vppClient) close() error {
	msgCodec := new(codec.MsgCodec)

	req := &memclnt.SockclntDelete{
		Index: c.clientIndex,
	}
	msg, err := msgCodec.EncodeMsg(req, c.sockDelMsgId)
	if err != nil {
		Log.Debugln("Encode error:", err)
		return err
	}
	// set non-0 context
	msg[5] = 124

	Log.Debugf("sending socklntDel (%d byes): % 0X\n", len(msg), msg)
	if err := c.write(msg); err != nil {
		Log.Debugln("Write error: ", err)
		return err
	}

	readDeadline := time.Now().Add(DisconnectTimeout)
	if err := c.conn.SetReadDeadline(readDeadline); err != nil {
		return err
	}
	msgReply, err := c.read()
	if err != nil {
		Log.Debugln("Read error:", err)
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			// we accept timeout for reply
			return nil
		}
		return err
	}
	// reset read deadline
	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	reply := new(memclnt.SockclntDeleteReply)
	if err := msgCodec.DecodeMsg(msgReply, reply); err != nil {
		Log.Debugln("Decode error:", err)
		return err
	}

	Log.Debugf("SockclntDeleteReply: Response=%v", reply.Response)

	return nil
}

func (c *vppClient) GetMsgID(msgName string, msgCrc string) (uint16, error) {
	msg := msgName + "_" + msgCrc
	msgID, ok := c.msgTable[msg]
	if !ok {
		return 0, fmt.Errorf("unknown message: %q", msg)
	}
	return msgID, nil
}

type reqHeader struct {
	//MsgID uint16
	ClientIndex uint32
	Context     uint32
}

func (c *vppClient) SendMsg(context uint32, data []byte) error {
	h := &reqHeader{
		ClientIndex: c.clientIndex,
		Context:     context,
	}
	buf := new(bytes.Buffer)
	if err := struc.Pack(buf, h); err != nil {
		return err
	}
	copy(data[2:], buf.Bytes())

	Log.Debugf("SendMsg (%d) context=%v client=%d: data: % 02X", len(data), context, c.clientIndex, data)

	if err := c.write(data); err != nil {
		Log.Debugln("write error: ", err)
		return err
	}

	return nil
}

func (c *vppClient) write(msg []byte) error {
	h := &msgheader{
		Data_len: uint32(len(msg)),
	}
	buf := new(bytes.Buffer)
	if err := struc.Pack(buf, h); err != nil {
		return err
	}
	header := buf.Bytes()

	// we lock to prevent mixing multiple message sends
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var w io.Writer = c.conn

	if n, err := w.Write(header); err != nil {
		return err
	} else {
		Log.Debugf(" - header sent (%d/%d): % 0X", n, len(header), header)
	}
	if n, err := w.Write(msg); err != nil {
		return err
	} else {
		Log.Debugf(" - msg sent (%d/%d): % 0X", n, len(msg), msg)
	}

	return nil
}

type msgHeader struct {
	MsgID   uint16
	Context uint32
}

func (c *vppClient) readerLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.quit:
			Log.Debugf("reader quit")
			return
		default:
		}

		msg, err := c.read()
		if err != nil {
			if isClosedError(err) {
				return
			}
			Log.Debugf("READ FAILED: %v", err)
			continue
		}
		h := new(msgHeader)
		if err := struc.Unpack(bytes.NewReader(msg), h); err != nil {
			Log.Debugf("unpacking header failed: %v", err)
			continue
		}

		Log.Debugf("recvMsg (%d) msgID=%d context=%v", len(msg), h.MsgID, h.Context)
		c.cb(h.MsgID, msg)
	}
}

type msgheader struct {
	Q                 int    `struc:"uint64"`
	Data_len          uint32 `struc:"uint32"`
	Gc_mark_timestamp uint32 `struc:"uint32"`
	//data              [0]uint8
}

func (c *vppClient) read() ([]byte, error) {
	Log.Debug("reading next msg..")

	header := make([]byte, 16)

	n, err := io.ReadAtLeast(c.reader, header, 16)
	if err != nil {
		return nil, err
	} else if n == 0 {
		Log.Debugln("zero bytes header")
		return nil, nil
	}
	if n != 16 {
		Log.Debug("invalid header data (%d): % 0X", n, header[:n])
		return nil, fmt.Errorf("invalid header (expected 16 bytes, got %d)", n)
	}
	Log.Debugf(" - read header %d bytes: % 0X", n, header)

	h := &msgheader{}
	if err := struc.Unpack(bytes.NewReader(header[:]), h); err != nil {
		return nil, err
	}
	Log.Debugf(" - decoded header: %+v", h)

	msgLen := int(h.Data_len)
	msg := make([]byte, msgLen)

	n, err = c.reader.Read(msg)
	if err != nil {
		return nil, err
	}
	Log.Debugf(" - read msg %d bytes (%d buffered)", n, c.reader.Buffered())

	if msgLen > n {
		remain := msgLen - n
		Log.Debugf("continue read for another %d bytes", remain)
		view := msg[n:]

		for remain > 0 {

			nbytes, err := c.reader.Read(view)
			if err != nil {
				return nil, err
			} else if nbytes == 0 {
				return nil, fmt.Errorf("zero nbytes")
			}

			remain -= nbytes
			Log.Debugf("another data received: %d bytes (remain: %d)", nbytes, remain)

			view = view[nbytes:]
		}
	}

	return msg, nil
}

func isClosedError(err error) bool {
	if err == io.EOF {
		return true
	}
	return strings.HasSuffix(err.Error(), "use of closed network connection")
}
