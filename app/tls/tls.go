package tls

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"

	acl "openstackcore-rdtagent/util/acl"
)

var signatureRWM sync.RWMutex
var adminCertSignature []string

func GetCertPool(cafile string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	// Should we get SystemCertPool ?
	data, err := ioutil.ReadFile(cafile)
	if err != nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM(data)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	return pool, nil
}

func NewAdminCertSignatures() ([]string, error) {
	signatures := []string{}
	files, err := acl.GetAdminCerts()
	if err != nil {
		return signatures, err
	}

	for _, f := range files {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			log.Errorf("Unable to read signatures file: %s. Error: %s", f, err)
		}
		block, _ := pem.Decode(dat)
		if block == nil || block.Type != "CERTIFICATE" {
			log.Errorf("Failed to decode client certificate %s. Certificate type: %s", f, block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Errorf("Failed to parse client certificate %s. Error: %s", f, err)
		} else {
			signatures = append(signatures, string(cert.Signature))
		}
	}

	return signatures, nil
}

func InitAdminCertSignatures() (err error) {
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// No choice to close watcher. V2 will support goroutine gracefully exit.
	// defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Create+fsnotify.Write+fsnotify.Remove) > 0 {
					log.Infof("Client cert files are changed, reload. Event: %s", event)
					cs, err := NewAdminCertSignatures()
					if err != nil {
						log.Errorf("Error to get client signatures list. %s", err)
					}
					signatureRWM.Lock()
					adminCertSignature = cs
					signatureRWM.Unlock()
					log.Infof("Load %d valid certificate signatures.", len(adminCertSignature))
				}
			case err := <-watcher.Errors:
				log.Errorf("Error to watch client certificate path. Error: %s", err)
			}
		}
	}()

	for _, p := range acl.GetCertsPath() {
		err = watcher.Add(p)
		if err != nil {
			return err
		}
	}
	adminCertSignature, err = NewAdminCertSignatures()
	return
}

func GetAdminCertSignatures() []string {
	// NOTE adminCertSignature should not be get once
	signatureRWM.RLock()
	defer signatureRWM.RUnlock()
	return adminCertSignature
}
