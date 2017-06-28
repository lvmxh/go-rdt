package app

import (
	"fmt"
	"log"
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

	if s.Addr == "" {
		s.Addr = appconfig.Def.Address
	}

	if s.Port == "" {
		s.Port = strconv.FormatUint(uint64(appconfig.Def.Port), 10)
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
	var d db.DB

	if c.Generic.DBBackend == "bolt" {
		d = new(db.BoltDB)
	} else if c.Generic.DBBackend == "mgo" {
		d = new(db.MgoDB)
	} else {
		return nil, fmt.Errorf("Unsupported DB backend %s", c.Generic.DBBackend)
	}

	err := d.Initialize(c.Generic.Transport, c.Generic.DBName)

	if err != nil {
		return nil, err
	}
	return d, nil
}

// Initialize server from config
func Initialize(c *Config) (*http.Server, error) {
	db, err := InitializeDB(c)

	if err != nil {
		return nil, err
	}

	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})

	cap := v1.CapabilitiesResource{}
	caches := v1.CachesResource{}
	policy := v1.PolicyResource{}
	wls := v1.WorkLoadResource{Db: db}

	// Register controller to container
	cap.Register(wsContainer)
	caches.Register(wsContainer)
	policy.Register(wsContainer)
	wls.Register(wsContainer)

	// Install adds the SgaggerUI webservices
	c.Swagger.WebServices = wsContainer.RegisteredWebServices()
	swagger.RegisterSwaggerService(*(c.Swagger), wsContainer)

	// TODO error handle
	return &http.Server{Addr: c.Generic.APIServerServiceIP + ":" + c.Generic.APIServerServicePort, Handler: wsContainer}, nil
}

// RunServer uses the provided options to run the apiserver.
func RunServer(s *options.ServerRunOptions) {

	config := BuildServerConfig(s)
	server, err := Initialize(config)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal(server.ListenAndServe())
	}
}
