package instana

import (
	"runtime"
	"time"

	"github.com/instana/go-sensor/acceptor"
)

// SnapshotS struct to hold snapshot data
type SnapshotS acceptor.RuntimeInfo

// MemoryS struct to hold snapshot data
type MemoryS acceptor.MemoryStats

// MetricsS struct to hold snapshot data
type MetricsS acceptor.Metrics

// EntityData struct to hold snapshot data
type EntityData acceptor.GoProcessData

type metricSender interface {
	Ready() bool
	SendMetrics(acceptor.Metrics) error
}

type meterS struct {
	numGC uint32

	logger LeveledLogger
	agent  metricSender
}

func newMeter(agent metricSender, logger LeveledLogger) *meterS {
	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing meter")

	meter := &meterS{
		logger: logger,
		agent:  agent,
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if !meter.agent.Ready() {
				continue
			}

			go meter.agent.SendMetrics(meter.collectMetrics())
		}
	}()

	return meter
}

func (m *meterS) setLogger(l LeveledLogger) {
	m.logger = l
}

func (r *meterS) collectMemoryMetrics() acceptor.MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ret := acceptor.MemoryStats{
		Alloc:         memStats.Alloc,
		TotalAlloc:    memStats.TotalAlloc,
		Sys:           memStats.Sys,
		Lookups:       memStats.Lookups,
		Mallocs:       memStats.Mallocs,
		Frees:         memStats.Frees,
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		HeapIdle:      memStats.HeapIdle,
		HeapInuse:     memStats.HeapInuse,
		HeapReleased:  memStats.HeapReleased,
		HeapObjects:   memStats.HeapObjects,
		PauseTotalNs:  memStats.PauseTotalNs,
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction}

	if r.numGC < memStats.NumGC {
		ret.PauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
		r.numGC = memStats.NumGC
	}

	return ret
}

func (r *meterS) collectMetrics() acceptor.Metrics {
	return acceptor.Metrics{
		CgoCall:     runtime.NumCgoCall(),
		Goroutine:   runtime.NumGoroutine(),
		MemoryStats: r.collectMemoryMetrics(),
	}
}
