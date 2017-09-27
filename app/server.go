package app

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	appConf "openstackcore-rdtagent/app/config"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	log "github.com/sirupsen/logrus"
	"openstackcore-rdtagent/api/v1"
	"openstackcore-rdtagent/db"
	"openstackcore-rdtagent/util/acl"
	"openstackcore-rdtagent/util/options"
	"openstackcore-rdtagent/util/auth"
)

type GenericConfig struct {
	APIServerServiceIP   string
	APIServerServicePort string
	EnableUISupport      bool
	DBBackend            string
	Transport            string
	DBName               string
}

// RDAgent config
type Config struct {
	Generic *GenericConfig
	Swagger *swagger.Config
}

// build RDAgent configuration from command line and configure file
func BuildServerConfig(s *options.ServerRunOptions) *Config {
	// FIXME (cmd line options does not override the config file options)
	appconfig := appConf.NewConfig()

	s.UnixSock = appconfig.Def.UnixSock

	if s.Addr == "" {
		s.Addr = appconfig.Def.Address
	}

	if s.Port == "" {
		s.Port = strconv.FormatUint(uint64(appconfig.Def.Port), 10)
	}

	if appconfig.Def.TLSPort != 0 {
		s.TLSPort = strconv.FormatUint(uint64(appconfig.Def.TLSPort), 10)
	}

	genericconfig := GenericConfig{
		APIServerServiceIP:   s.Addr,
		APIServerServicePort: s.Port,
		DBBackend:            appconfig.Db.Backend,
		Transport:            appconfig.Db.Transport,
		DBName:               appconfig.Db.DBName,
	}

	swaggerconfig := swagger.Config{
		WebServicesUrl: fmt.Sprintf("http://%s:%s", s.Addr, s.Port),
		ApiPath:        "/apidocs.json",
		SwaggerPath:    "/apidocs/", // Optionally, specifiy where the UI is located
		// FIXME (eliqiao): this depends on https://github.com/swagger-api/swagger-ui.git need to copy dist from it
		SwaggerFilePath: "/usr/local/share/go/src/github.com/wordnik/swagger-ui/dist",
		ApiVersion:      "1.0",
	}

	return &Config{
		Generic: &genericconfig,
		Swagger: &swaggerconfig,
	}
}

func InitializeDB(c *Config) (db.DB, error) {
	return db.NewDB()
	// no need Initialize. We can Initialize it at bootcheck
	// also, Initialize should be for the whole DB setting.
	// d.Initialize(c.Generic.Transport, c.Generic.DBName)
}

func TlsACL(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	appconf := appConf.NewConfig()
	if req.Request.TLS == nil || appConf.ClientAuth[appconf.Def.ClientAuth] < appConf.ClientAuth["challenge_given"] {
		chain.ProcessFilter(req, resp)
		return
	}

	ou := ""
	for _, v := range req.Request.TLS.PeerCertificates[0].Subject.OrganizationalUnit {
		if strings.ToLower(v) == "admin" {
			ou = "admin"
			break
		}
		if strings.ToLower(v) == "user" {
			ou = "user"
		}
	}
	cn := req.Request.TLS.PeerCertificates[0].Subject.CommonName
	user := cn
	if ou != "" {
		user = ou
	}
	e, _ := acl.NewEnforcer()
	if e.Enforce(req, user) != true {
		log.Errorf("User is not allow to access this resource")
		resp.WriteErrorString(401, cn+" is not Authorized")
		return
	}
	chain.ProcessFilter(req, resp)
}

// Initialize server from config
func Initialize(c *Config) (*restful.Container, error) {
	db, err := InitializeDB(c)
	if err != nil {
		return nil, err
	}

	wsContainer := restful.NewContainer()
	wsContainer.Filter(TlsACL)
	wsContainer.Filter(auth.PamAuthenticate)
	wsContainer.Router(restful.CurlyRouter{})

	cap := v1.CapabilitiesResource{}
	caches := v1.CachesResource{}
	policy := v1.PolicyResource{}
	hospitality := v1.HospitalityResource{}
	wls := v1.WorkLoadResource{Db: db}

	// Register controller to container
	cap.Register(wsContainer)
	caches.Register(wsContainer)
	policy.Register(wsContainer)
	hospitality.Register(wsContainer)
	wls.Register(wsContainer)

	// Install adds the SgaggerUI webservices
	c.Swagger.WebServices = wsContainer.RegisteredWebServices()
	swagger.RegisterSwaggerService(*(c.Swagger), wsContainer)

	// TODO error handle
	return wsContainer, nil
}

// TODO an individual go file for TLS. And move these functions to this file.
func getCertPool(cafile string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	// Should we get SystemCertPool ?
	data, err := ioutil.ReadFile(cafile)
	if err != nil {
		return nil, err
	}
	ok := pool.AppendCertsFromPEM(data)
	if !ok {
		return nil, errors.New("failed to parse root certificate!")
	}
	return pool, nil
}

func genTLSConfig() (*tls.Config, error) {
	var roots *x509.CertPool
	var clientPool *x509.CertPool
	tlsfiles := map[string]string{}
	appconf := appConf.NewConfig()
	files, err := filepath.Glob(appconf.Def.CertPath + "/*.pem")
	if err != nil {
		return nil, err
	}
	// avoid to check whether files exist.
	for _, f := range files {
		switch filepath.Base(f) {
		case appConf.CAFile:
			tlsfiles["ca"] = f
			roots, err = getCertPool(f)
			if err != nil {
				return nil, err
			}
		case appConf.CertFile:
			tlsfiles["cert"] = f
		case appConf.KeyFile:
			tlsfiles["key"] = f
		}
	}
	if len(tlsfiles) < 3 {
		missing := []string{}
		for _, k := range []string{"cert", "ca", "key"} {
			_, ok := tlsfiles[k]
			if !ok {
				missing = append(missing, k)
			}
		}
		return nil, fmt.Errorf("Missing enough files for tls config: %s.", strings.Join(missing, ", "))
	}

	// In product env, ClientAuth should >= challenge_given
	clientauth, ok := appConf.ClientAuth[appconf.Def.ClientAuth]
	if !ok {
		return nil, errors.New(
			"Unknow ClientAuth config setting: " + appconf.Def.CertPath)
	}
	if appConf.ClientAuth[appconf.Def.ClientAuth] >= appConf.ClientAuth["challenge_given"] {
		clientPool, err = getCertPool(filepath.Join(appconf.Def.ClientCAPath, appConf.ClientCAFile))
		if err != nil {
			return nil, err
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(tlsfiles["cert"], tlsfiles["key"])
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      roots,
		ClientAuth:   clientauth,
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    clientPool}, nil
}

// RunServer uses the provided options to run the apiserver.
func RunServer(s *options.ServerRunOptions) {

	var server *http.Server
	config := BuildServerConfig(s)
	container, err := Initialize(config)
	if err != nil {
		log.Fatal(err)
	}

	if s.TLSPort == "" {
		server = &http.Server{
			Addr:    s.Addr + ":" + s.Port,
			Handler: container}
	} else {
		// TODO We need to config server.TLSConfig
		// TODO Support self-sign CA. self-sign CA can be in development evn.
		tlsconf, err := genTLSConfig()
		if err != nil {
			log.Fatal(err)
		}

		server = &http.Server{
			Addr:      s.Addr + ":" + s.TLSPort,
			Handler:   container,
			TLSConfig: tlsconf}
	}

	server_start := func() {
		if s.TLSPort == "" {
			log.Fatal(server.ListenAndServe())
		} else {
			log.Fatal(server.ListenAndServeTLS("", "")) // Use certs from TLSConfig.
		}
	}
	if s.UnixSock == "" {
		server_start()
	} else {
		go func() {
			server_start()
		}()
	}

	// Unix Socket.
	config = BuildServerConfig(s)
	container, err = Initialize(config)
	if err != nil {
		log.Fatal(err)
	}

	userver := &http.Server{
		Handler: container}

	unixListener, err := net.Listen("unix", s.UnixSock)
	if err != nil {
		log.Info(err, unixListener)
		return
	}
	// TODO need to check, should defer unixListener.Close()
	defer func() {
		if unixListener != nil {
			log.Infof("Close Unix socket listener. RMD exits!")
			unixListener.Close()
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	go func(l net.Listener, c chan os.Signal) {
		sig := <-c
		if l != nil {
			log.Infof("Close Unix socket listener.")
			l.Close()
		}
		log.Infof("Caught signal %s: RMD exits!", sig)
		os.Exit(0)
	}(unixListener, sigchan)

	//REMOVE these 2 line codes, if we want to support Unix Socket!
	unixListener.Close()
	log.Fatal("Sorry, do not support Unix listener at present!")

	err = userver.Serve(unixListener)
	if err != nil {
		log.Fatal(err)
	}
}
