package statsd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/atlassian/gostatsd/types"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// ip packet size is stored in two bytes and that is how big in theory the packet can be.
// In practice it is highly unlikely but still possible to get packets bigger than usual MTU of 1500.
const packetSizeUDP = 0xffff

// Handler interface can be used to handle metrics and events for a Receiver.
type Handler interface {
	DispatchMetric(context.Context, *types.Metric) error
	DispatchEvent(context.Context, *types.Event) error
}

// RunableHandler extends Handler interface to add the Run method that needs to be executed in a separate
// goroutine to make the handler work.
type RunableHandler interface {
	Handler
	Run(context.Context) error
}

// Receiver receives data on its PacketConn and converts lines into Metrics.
// For each types.Metric it calls Handler.HandleMetric()
type Receiver interface {
	Receive(context.Context, net.PacketConn) error
	GetStats() ReceiverStats
}

// ReceiverStats holds statistics for a Receiver.
type ReceiverStats struct {
	LastPacket      time.Time
	BadLines        uint64
	PacketsReceived uint64
	MetricsReceived uint64
	EventsReceived  uint64
}

type metricReceiver struct {
	// Counter fields below must be read/written only using atomic instructions.
	// 64-bit fields must be the first fields in the struct to guarantee proper memory alignment.
	// See https://golang.org/pkg/sync/atomic/#pkg-note-BUG
	lastPacket      int64 // When last packet was received. Unix timestamp in nsec.
	badLines        uint64
	packetsReceived uint64
	metricsReceived uint64
	eventsReceived  uint64
	handler         Handler    // handler to invoke
	namespace       string     // Namespace to prefix all metrics
	tags            types.Tags // Tags to add to all metrics
}

// NewMetricReceiver initialises a new Receiver.
func NewMetricReceiver(ns string, tags []string, handler Handler) Receiver {
	return &metricReceiver{
		handler:   handler,
		namespace: ns,
		tags:      tags,
	}
}

// GetStats returns current Receiver stats. Safe for concurrent use.
func (mr *metricReceiver) GetStats() ReceiverStats {
	return ReceiverStats{
		LastPacket:      time.Unix(0, atomic.LoadInt64(&mr.lastPacket)),
		BadLines:        atomic.LoadUint64(&mr.badLines),
		PacketsReceived: atomic.LoadUint64(&mr.packetsReceived),
		MetricsReceived: atomic.LoadUint64(&mr.metricsReceived),
		EventsReceived:  atomic.LoadUint64(&mr.eventsReceived),
	}
}

// Receive accepts incoming datagrams on c, parses them and calls Handler.DispatchMetric() for each metric
// and Handler.DispatchEvent() for each event.
func (mr *metricReceiver) Receive(ctx context.Context, c net.PacketConn) error {
	buf := make([]byte, packetSizeUDP)
	for {
		// This will error out when the socket is closed.
		nbytes, addr, err := c.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && !netErr.Temporary() {
				select {
				case <-ctx.Done():
				default:
					return fmt.Errorf("non-temporary error reading from socket: %v", err)
				}
				return nil
			}
			log.Warnf("Error reading from socket: %v", err)
			continue
		}
		// TODO consider updating counter for every N-th iteration to reduce contention
		atomic.AddUint64(&mr.packetsReceived, 1)
		atomic.StoreInt64(&mr.lastPacket, time.Now().UnixNano())
		if err := mr.handlePacket(ctx, addr, buf[:nbytes]); err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return err
			}
			log.Warnf("Failed to handle packet: %v", err)
		}
	}
}

// handlePacket handles the contents of a datagram and calls Handler.DispatchMetric()
// for each line that successfully parses into a types.Metric and Handler.DispatchEvent() for each event.
func (mr *metricReceiver) handlePacket(ctx context.Context, addr net.Addr, msg []byte) error {
	var numMetrics, numEvents uint16
	var exitError error
	var ip types.IP
	buf := bytes.NewBuffer(msg)
	addrStr := addr.String()
	n := strings.LastIndexByte(addrStr, ':')
	if n <= 0 {
		// This should not ever happen.
		log.Errorf("Cannot parse source address %q", addrStr)
	} else {
		ip = types.IP(addrStr[:n])
	}
	for {
		line, readerr := buf.ReadBytes('\n')

		// protocol does not require line to end in \n, if EOF use received line if valid
		if readerr != nil && readerr != io.EOF {
			exitError = fmt.Errorf("error reading message from %s: %v", addr, readerr)
			break
		} else if readerr != io.EOF {
			// remove newline, only if not EOF
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		}

		if len(line) > 1 {
			metric, event, err := mr.parseLine(line)
			if err != nil {
				// logging as debug to avoid spamming logs when a bad actor sends
				// badly formatted messages
				log.Debugf("Error parsing line %q from %s: %v", line, addr, err)
				atomic.AddUint64(&mr.badLines, 1)
				continue
			}
			if metric != nil {
				numMetrics++
				metric.Tags = append(metric.Tags, mr.tags...)
				metric.SourceIP = ip
				err = mr.handler.DispatchMetric(ctx, metric)
			} else if event != nil {
				numEvents++
				event.Tags = append(event.Tags, mr.tags...)
				event.SourceIP = ip
				if event.DateHappened == 0 {
					event.DateHappened = time.Now().Unix()
				}
				err = mr.handler.DispatchEvent(ctx, event)
			} else {
				// Should never happen.
				log.Panic("Both event and metric are nil")
			}
			if err != nil {
				exitError = err
				break
			}
		}

		if readerr == io.EOF {
			// if was EOF, finished handling
			break
		}
	}
	atomic.AddUint64(&mr.metricsReceived, uint64(numMetrics))
	atomic.AddUint64(&mr.eventsReceived, uint64(numEvents))
	return exitError
}

// parseLine with lexer impl.
func (mr *metricReceiver) parseLine(line []byte) (*types.Metric, *types.Event, error) {
	l := lexer{}
	return l.run(line, mr.namespace)
}
