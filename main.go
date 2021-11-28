package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kardianos/service"
	"gopkg.in/yaml.v3"

	"github.com/vrherog/daemonupgrader/version"
)

var (
	appVersion bool
	svcFlag    string

	logger service.Logger

	upgradeReadyFile = `upgrade.ready`
	upgradeOkFile    = `upgrade.ok`
)

type PackageStatus uint8

const (
	checkServiceStatus PackageStatus = iota
	checkUpgrade
	upgradeOk
)

type ServiceConfig struct {
	Name             string                 `yaml:"name"`
	DisplayName      string                 `yaml:"displayName,omitempty"`
	Description      string                 `yaml:"description,omitempty"`
	Arguments        []string               `yaml:"arguments,omitempty"`
	UserName         string                 `yaml:"username,omitempty"`
	WorkingDirectory string                 `yaml:"workDirectory,omitempty"`
	Options          map[string]interface{} `yaml:"options,omitempty"`
	Services         []ServiceInfo          `yaml:"services"`
	Packages         []UpgradePackageInfo   `yaml:"packages"`
}

func init() {
	flag.StringVar(&svcFlag, `service`, "", `Control the system service.`)
	flag.BoolVar(&appVersion, "version", false, "show version")
	flag.BoolVar(&appVersion, "v", false, "show version")
}

func main() {
	flag.Parse()

	if appVersion {
		fmt.Println(version.BuildVersion)
		return
	}

	var err error
	var confFile string
	var conf ServiceConfig
	var execFile, _ = os.Executable()
	var execName = filepath.Base(execFile)
	var ext = filepath.Ext(execName)
	if len(ext) > 0 {
		execName = execName[:len(execName)-len(ext)]
	}
	var execDir = filepath.Dir(execFile)
	for _, confFile = range []string{
		`config.yaml`,
		fmt.Sprintf(`%s.yaml`, execName),
		fmt.Sprintf(`%s.conf`, execName),
	} {
		confFile = filepath.Join(execDir, confFile)
		if _, err = os.Stat(confFile); err == nil || os.IsExist(err) {
			var content []byte
			content, err = ioutil.ReadFile(confFile)
			if err != nil {
				log.Fatal(err)
				return
			}
			err = yaml.Unmarshal(content, &conf)
			if err != nil {
				log.Fatal(err)
				return
			}
			break
		}
	}
	if conf.Name == "" {
		log.Fatal(errors.New(`config file not exist or invalid`))
		return
	}

	upgradeReadyFile = filepath.Join(execDir, upgradeReadyFile)
	upgradeOkFile = filepath.Join(execDir, upgradeOkFile)

	var svcConfig = &service.Config{
		Name:             conf.Name,
		DisplayName:      conf.DisplayName,
		Description:      conf.Description,
		UserName:         conf.UserName,
		Arguments:        conf.Arguments,
		WorkingDirectory: conf.WorkingDirectory,
	}
	if conf.Name == `` {
		conf.Name = execName
	}
	if conf.Services != nil {
		for _, s := range conf.Services {
			if s.Interval.Seconds() < 3 {
				s.Interval = time.Second * 3
			}
		}
	}
	if conf.Packages != nil {
		for _, p := range conf.Packages {
			if p.Interval.Seconds() < 30 {
				p.Interval = time.Minute * 30
			}
		}
	}

	var prg = &program{
		services: conf.Services,
		packages: conf.Packages,
		tasks:    sync.Map{},
	}
	var srv service.Service
	srv, err = service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	var errs = make(chan error, 5)
	if logger, err = srv.Logger(errs); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			var err = <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if svcFlag != `` {
		if err = service.Control(srv, svcFlag); err != nil {
			log.Printf(`Valid actions: %q\n`, service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	if err = srv.Run(); err != nil {
		_ = logger.Error(err)
	}
}
