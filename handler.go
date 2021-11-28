package main

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/service"

	"github.com/vrherog/daemonupgrader/utils"
	"github.com/vrherog/daemonupgrader/version"
)

type UpgradeReadyInfo struct {
	WorkDirectory string `json:"workDirectory"`
	PackageDir    string `json:"package_dir"`
	Version       string `json:"version"`
}

func (p *program) checkServiceStatus(name string) {
	p.tasks.Store(name, checkServiceStatus)
	if srv, err := service.New(&program{}, &service.Config{Name: name}); err == nil {
		if status, err := srv.Status(); err == nil {
			if status == service.StatusStopped {
				if err = srv.Start(); err != nil {
					_ = logger.Errorf(`start service %s %s`, name, err)
				}
			}
		}
	}
	p.tasks.Delete(name)
}

func (p *program) checkUpgrade(packageInfo UpgradePackageInfo) {
	p.tasks.Store(packageInfo.Name, checkUpgrade)
	var err error
	var dirInfo os.FileInfo
	if dirInfo, err = os.Stat(packageInfo.WorkDirectory); (err == nil || os.IsExist(err)) && dirInfo.IsDir() {
		var urlPackage *url.URL
		urlPackage, err = url.Parse(packageInfo.UriDownloadPackage)
		if err == nil {
			var parts = strings.Split(urlPackage.Path, `/`)
			if len(parts) > 0 {
				var filename = parts[len(parts)-1]
				var packageType = filepath.Ext(filename)
				var remoteVer, localVer string
				if remoteVer, _, err = utils.RequestText(packageInfo.UriCheckVersion, `GET`, ``); err == nil {
					if localVer, err = utils.ExecCommandString(packageInfo.CommandGetVersion); err == nil {
						if comp, ok := version.CompareVersion(remoteVer, strings.TrimSpace(string(localVer))); ok && comp == 1 {
							var upgradeReadyInfo map[string]UpgradeReadyInfo
							if ok = utils.ReadJsonFile(upgradeReadyFile, &upgradeReadyInfo); ok {
								upgradeReadyInfo = make(map[string]UpgradeReadyInfo)
							}
							var notReady = true
							if info, ok := upgradeReadyInfo[packageInfo.Name]; ok {
								if comp, ok := version.CompareVersion(info.Version, remoteVer); ok && comp >= 0 {
									notReady = false
								}
							}
							if notReady {
								var tempDir string
								tempDir, err = ioutil.TempDir(os.TempDir(), `upgrade`)
								if err != nil {
									tempDir = os.TempDir()
								}
								var packageFile = filepath.Join(os.TempDir(), filename)
								if _, err = utils.DownloadFile(packageFile, packageInfo.UriDownloadPackage, `GET`, ``); err == nil {
									_ = logger.Infof(`find new version: %s`, packageInfo.Name)
									switch packageType {
									case `.zip`:
										err = utils.Unzip(packageFile, tempDir)
									case `.gz`:
										err = utils.ExtractGzip(packageFile, tempDir)
									default:
										err = errors.New(`not supported type`)
									}
									if err == nil {
										_ = os.Remove(packageFile)
										_ = logger.Infof(`new version download completed:%s %s`, packageInfo.Name, tempDir)
										if packageInfo.NeedShutdown {
											upgradeReadyInfo[packageInfo.Name] = UpgradeReadyInfo{
												WorkDirectory: packageInfo.WorkDirectory,
												PackageDir:    tempDir,
												Version:       remoteVer,
											}
											if err = utils.WriteJsonFile(upgradeReadyFile, upgradeReadyInfo); err == nil {
												_ = logger.Infof(`upgrade ready: %s`, packageInfo.Name)
											}
										} else {
											if err = utils.CopyFiles(tempDir, packageInfo.WorkDirectory); err == nil {
												err = os.RemoveAll(tempDir)
												_ = logger.Infof(`upgrade completed: %s`, packageInfo.Name)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if err != nil {
		_ = logger.Error(err)
	}
	p.tasks.Delete(packageInfo.Name)
}

func (p *program) upgradePackage(name string) {
	p.tasks.Store(name, upgradeOk)
	var upgradeReadyInfo map[string]UpgradeReadyInfo
	var err error
	if ok := utils.ReadJsonFile(upgradeReadyFile, &upgradeReadyInfo); ok {
		if info, ok := upgradeReadyInfo[name]; ok {
			if err = utils.CopyFiles(info.PackageDir, info.WorkDirectory); err == nil {
				err = os.RemoveAll(info.PackageDir)
				if content, ok := utils.ReadTextFile(upgradeOkFile); ok {
					var buffer = make([]string, 0)
					for _, item := range strings.Fields(content) {
						if item != name {
							buffer = append(buffer, item)
						}
					}
					if len(buffer) > 0 {
						if err = utils.WriteTextFile(upgradeOkFile, strings.Join(buffer, `\n`)); err == nil {
							_ = logger.Infof(`upgrade completed: %s`, name)
						}
					} else {
						if err = os.Remove(upgradeOkFile); err == nil {
							_ = logger.Infof(`upgrade completed: %s`, name)
						}
					}
					delete(upgradeReadyInfo, name)
					if len(upgradeReadyInfo) > 0 {
						err = utils.WriteJsonFile(upgradeReadyFile, upgradeReadyInfo)
					} else {
						err = os.Remove(upgradeReadyFile)
					}
				}
			}
		}
	}
	if err != nil {
		_ = logger.Error(err)
	}
	p.tasks.Delete(name)
}
