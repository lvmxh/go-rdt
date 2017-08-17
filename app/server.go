package app

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strconv"

	appConf "openstackcore-rdtagent/app/config"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	"openstackcore-rdtagent/api/v1"
	"openstackcore-rdtagent/db"
	"openstackcore-rdtagent/util/options"
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

// Initialize server from config
func Initialize(c *Config) (*restful.Container, error) {
	db, err := InitializeDB(c)
	if err != nil {
		return nil, err
	}

	wsContainer := restful.NewContainer()
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
		log.Fatal("Sorry, do not support TLS listener at present!")
		server = &http.Server{
			Addr:    s.Addr + ":" + s.TLSPort,
			Handler: container}
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

	config = BuildServerConfig(s)
	container, err = Initialize(config)
	if err != nil {
		log.Fatal(err)
	}
	userver := &http.Server{
		Handler: container}
	// TODO need to check, should defer unixListener.Close()
	unixListener, err := net.Listen("unix", s.UnixSock)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(userver.Serve(unixListener))
}
