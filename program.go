package main

import (
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"

	"github.com/vrherog/daemonupgrader/utils"
)

type ServiceInfo struct {
	Name     string        `yaml:"name"`
	Interval time.Duration `yaml:"interval,omitempty"`
}

type UpgradePackageInfo struct {
	Name               string        `yaml:"name"`
	Interval           time.Duration `yaml:"interval,omitempty"`
	UriCheckVersion    string        `yaml:"uriCheckVersion"`
	UriDownloadPackage string        `yaml:"uriDownloadPackage"`
	WorkDirectory      string        `yaml:"workDirectory"`
	CommandGetVersion  string        `yaml:"commandGetVersion"`
	NeedShutdown       bool          `yaml:"needShutdown,omitempty"`
}

func (p *UpgradePackageInfo) Validate() bool {
	return p.UriCheckVersion != `` && p.UriDownloadPackage != `` && p.WorkDirectory != `` && p.CommandGetVersion != ``
}

type program struct {
	tick     uint64
	exit     chan struct{}
	tasks    sync.Map
	services []ServiceInfo
	packages []UpgradePackageInfo
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		_ = logger.Info(`Running in terminal.`)
	}
	p.exit = make(chan struct{})

	go p.run()
	return nil
}

func (p *program) run() error {
	var ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if content, ok := utils.ReadTextFile(upgradeOkFile); ok {
				for _, name := range strings.Fields(content) {
					if _, ok := p.tasks.Load(name); !ok {
						go p.upgradePackage(name)
					}
				}
			}
			for _, s := range p.services {
				if p.tick%uint64(s.Interval.Seconds()) == 0 {
					if _, ok := p.tasks.Load(s.Name); !ok {
						go p.checkServiceStatus(s.Name)
					}
				}
			}
			for _, s := range p.packages {
				if s.Validate() {
					if p.tick%uint64(s.Interval.Seconds()) == 0 {
						if v, ok := p.tasks.Load(s.Name); !ok || v == checkServiceStatus {
							go p.checkUpgrade(s)
						}
					}
				}
			}
			p.tick++
		case <-p.exit:
			ticker.Stop()
			return nil
		}
	}
}

func (p *program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}
