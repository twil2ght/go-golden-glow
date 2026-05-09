package tracer

import (
	"fmt"
	"goldenglow/pkg/log"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/tracer"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"
)

func init() {
	plugin.DefaultManager.Register(pluginName, New())
}

const pluginName = "tracer"

var logger = log.Default()

const (
	resetOnIdle     = false
	writeOnIdle     = false
	writeOnShutdown = true
)

type tracePlugin struct {
	collector   *tracer.Collector
	lastWritten int
}

func (p *tracePlugin) Init() {}

func (p *tracePlugin) Shutdown() {
	if writeOnShutdown && p.hasNewEvents() {
		p.writeFiles()
	}
}

func (p *tracePlugin) OnRegisterIdleHandler(mgr registry.Interface[runner.IdleHandler]) {
	mgr.Register(pluginName, func() bool {
		if writeOnIdle && p.hasNewEvents() {
			p.writeFiles()
		}
		if resetOnIdle {
			p.collector.Reset()
			p.lastWritten = 0
		}
		return true
	})
}

func (p *tracePlugin) hasNewEvents() bool {
	return p.collector.Len() > p.lastWritten
}

func (p *tracePlugin) OnRegisterTraceHandler(mgr registry.Interface[runner.TraceHandler]) {
	mgr.Register(pluginName, func(event tracer.Event) {
		p.collector.Record(event)
	})
}

func (p *tracePlugin) writeFiles() {
	if p.collector.Len() == 0 {
		return
	}

	outDir := filepath.Join(utils.RootDir, "trace")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		logger.Error("tracer: create output dir", "err", err)
		return
	}

	htmlPath := filepath.Join(outDir, "execution.html")
	if err := os.WriteFile(htmlPath, []byte(p.collector.HTML()), 0644); err != nil {
		logger.Error("tracer: write HTML", "err", err)
	} else {
		logger.Info("tracer: HTML graph written", "path", htmlPath)
	}

	jsonPath := filepath.Join(outDir, "execution.json")
	if err := os.WriteFile(jsonPath, []byte(p.collector.JSON()), 0644); err != nil {
		logger.Error("tracer: write JSON", "err", err)
	} else {
		logger.Info("tracer: JSON trace written", "path", jsonPath)
	}

	p.lastWritten = p.collector.Len()
	fmt.Printf("[tracer] wrote %d events → %s, %s\n", p.collector.Len(), htmlPath, jsonPath)
}

func New() plugin.Interface {
	return &tracePlugin{
		collector: &tracer.Collector{},
	}
}
