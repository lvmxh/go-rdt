package tls

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	appConf "openstackcore-rdtagent/app/config"
	acl "openstackcore-rdtagent/util/acl"
)

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
	appconf := appConf.NewConfig()

	// Only support one client ca at present.
	clientPool, err := GetCertPool(filepath.Join(appconf.Def.ClientCAPath, appConf.ClientCAFile))
	if err != nil {
		return signatures, err
	}

	opts := x509.VerifyOptions{
		Roots: clientPool,
	}

	for _, f := range files {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			return signatures, err
		}
		block, _ := pem.Decode(dat)
		if block == nil || block.Type != "CERTIFICATE" {
			return signatures, errors.New("failed to parse root certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return signatures, err
		}
		// FIXME support verify cert.
		fmt.Println(opts.DNSName)
		// if _, err := cert.Verify(opts); err != nil {
		// 	return signatures, err
		// }

		signatures = append(signatures, string(cert.Signature))
	}

	return signatures, nil
}

func InitAdminCertSignatures() (err error) {
	adminCertSignature, err = NewAdminCertSignatures()
	return
}

func GetAdminCertSignatures() []string {
	// NOTE adminCertSignature should not be get once
	return adminCertSignature
}
