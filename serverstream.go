package gortsplib

import (
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ogsts/gortsplib/pkg/base"
	"github.com/ogsts/gortsplib/pkg/liberrors"
)

type listenerPair struct {
	rtpListener  *serverUDPListener
	rtcpListener *serverUDPListener
}

type trackInfo struct {
	lastSequenceNumber uint32
	lastTimeRTP        uint32
	lastTimeNTP        int64
	lastSSRC           uint32
}

// ServerStream represents a single stream.
// This is in charge of
// - distributing the stream to each reader
// - allocating multicast listeners
// - gathering infos about the stream to generate SSRC and RTP-Info
type ServerStream struct {
	s      *Server
	tracks Tracks

	mutex              sync.RWMutex
	readersUnicast     map[*ServerSession]struct{}
	readers            map[*ServerSession]struct{}
	multicastListeners []*listenerPair
	trackInfos         []*trackInfo
}

// NewServerStream allocates a ServerStream.
func NewServerStream(tracks Tracks) *ServerStream {
	st := &ServerStream{
		readersUnicast: make(map[*ServerSession]struct{}),
		readers:        make(map[*ServerSession]struct{}),
	}

	st.tracks = cloneAndClearTracks(tracks)

	st.trackInfos = make([]*trackInfo, len(tracks))
	for i := range st.trackInfos {
		st.trackInfos[i] = &trackInfo{}
	}

	return st
}

// Close closes a ServerStream.
func (st *ServerStream) Close() error {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.s != nil {
		select {
		case st.s.streamRemove <- st:
		case <-st.s.ctx.Done():
		}
	}

	for ss := range st.readers {
		ss.Close()
	}

	if st.multicastListeners != nil {
		for _, l := range st.multicastListeners {
			l.rtpListener.close()
			l.rtcpListener.close()
		}
		st.multicastListeners = nil
	}

	st.readers = nil
	st.readersUnicast = nil

	return nil
}

// Tracks returns the tracks of the stream.
func (st *ServerStream) Tracks() Tracks {
	return st.tracks
}

func (st *ServerStream) ssrc(trackID int) uint32 {
	return atomic.LoadUint32(&st.trackInfos[trackID].lastSSRC)
}

func (st *ServerStream) timestamp(trackID int) uint32 {
	lastTimeRTP := atomic.LoadUint32(&st.trackInfos[trackID].lastTimeRTP)
	lastTimeNTP := atomic.LoadInt64(&st.trackInfos[trackID].lastTimeNTP)
	clockRate, _ := st.tracks[trackID].ClockRate()

	if lastTimeRTP == 0 || lastTimeNTP == 0 {
		return 0
	}

	return uint32(uint64(lastTimeRTP) +
		uint64(time.Since(time.Unix(lastTimeNTP, 0)).Seconds()*float64(clockRate)))
}

func (st *ServerStream) lastSequenceNumber(trackID int) uint16 {
	return uint16(atomic.LoadUint32(&st.trackInfos[trackID].lastSequenceNumber))
}

func (st *ServerStream) readerAdd(
	ss *ServerSession,
	protocol base.StreamProtocol,
	delivery base.StreamDelivery,
	clientPorts *[2]int,
) error {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if st.s == nil {
		st.s = ss.s
		select {
		case st.s.streamAdd <- st:
		case <-st.s.ctx.Done():
		}
	}

	// if new reader is a UDP-unicast reader, check that its port are not already
	// in use by another reader.
	if protocol == base.StreamProtocolUDP && delivery == base.StreamDeliveryUnicast {
		for r := range st.readersUnicast {
			if *r.setuppedProtocol == base.StreamProtocolUDP &&
				*r.setuppedDelivery == base.StreamDeliveryUnicast &&
				r.ip().Equal(ss.ip()) &&
				r.zone() == ss.zone() {
				for _, rt := range r.setuppedTracks {
					if rt.udpRTPPort == clientPorts[0] {
						return liberrors.ErrServerUDPPortsAlreadyInUse{Port: rt.udpRTPPort}
					}
				}
			}
		}
	}

	// allocate multicast listeners
	if protocol == base.StreamProtocolUDP &&
		delivery == base.StreamDeliveryMulticast &&
		st.multicastListeners == nil {
		st.multicastListeners = make([]*listenerPair, len(st.tracks))

		for i := range st.tracks {
			rtpListener, rtcpListener, err := newServerUDPListenerMulticastPair(st.s)
			if err != nil {
				for _, l := range st.multicastListeners {
					if l != nil {
						l.rtpListener.close()
						l.rtcpListener.close()
					}
				}
				st.multicastListeners = nil
				return err
			}

			st.multicastListeners[i] = &listenerPair{
				rtpListener:  rtpListener,
				rtcpListener: rtcpListener,
			}
		}
	}

	st.readers[ss] = struct{}{}

	return nil
}

func (st *ServerStream) readerRemove(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	delete(st.readers, ss)

	if len(st.readers) == 0 && st.multicastListeners != nil {
		for _, l := range st.multicastListeners {
			l.rtpListener.close()
			l.rtcpListener.close()
		}
		st.multicastListeners = nil
	}
}

func (st *ServerStream) readerSetActive(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if *ss.setuppedDelivery == base.StreamDeliveryUnicast {
		st.readersUnicast[ss] = struct{}{}
	} else {
		for trackID := range ss.setuppedTracks {
			st.multicastListeners[trackID].rtcpListener.addClient(
				ss.ip(), st.multicastListeners[trackID].rtcpListener.port(), ss, trackID, false)
		}
	}
}

func (st *ServerStream) readerSetInactive(ss *ServerSession) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if *ss.setuppedDelivery == base.StreamDeliveryUnicast {
		delete(st.readersUnicast, ss)
	} else if st.multicastListeners != nil {
		for trackID := range ss.setuppedTracks {
			st.multicastListeners[trackID].rtcpListener.removeClient(ss)
		}
	}
}

// WriteFrame writes a frame to all the readers of the stream.
func (st *ServerStream) WriteFrame(trackID int, streamType StreamType, payload []byte) {
	if streamType == StreamTypeRTP && len(payload) >= 8 {
		track := st.trackInfos[trackID]

		sequenceNumber := binary.BigEndian.Uint16(payload[2:4])
		atomic.StoreUint32(&track.lastSequenceNumber, uint32(sequenceNumber))

		timestamp := binary.BigEndian.Uint32(payload[4:8])
		atomic.StoreUint32(&track.lastTimeRTP, timestamp)
		atomic.StoreInt64(&track.lastTimeNTP, time.Now().Unix())

		ssrc := binary.BigEndian.Uint32(payload[8:12])
		atomic.StoreUint32(&track.lastSSRC, ssrc)
	}

	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// send unicast
	for r := range st.readersUnicast {
		r.WriteFrame(trackID, streamType, payload)
	}

	// send multicast
	if st.multicastListeners != nil {
		if streamType == StreamTypeRTP {
			st.multicastListeners[trackID].rtpListener.write(payload, &net.UDPAddr{
				IP:   st.multicastListeners[trackID].rtpListener.ip(),
				Zone: "",
				Port: st.multicastListeners[trackID].rtpListener.port(),
			})
		} else {
			st.multicastListeners[trackID].rtcpListener.write(payload, &net.UDPAddr{
				IP:   st.multicastListeners[trackID].rtpListener.ip(),
				Zone: "",
				Port: st.multicastListeners[trackID].rtcpListener.port(),
			})
		}
	}
}
