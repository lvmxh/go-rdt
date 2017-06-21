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
	"openstackcore-rdtagent/pkg/db"
	"openstackcore-rdtagent/pkg/model/capabilities"
	"openstackcore-rdtagent/util/options"
)

// TODO move this out of app server
type GenericAPIConfig struct {
	//APIServerServiceIP   net.IP
	APIServerServiceIP   string
	APIServerServicePort string
	EnableUISupport      bool
	Transport            string
}

// TODO move this out of app server
// This is server config, include any running status data in this struct
// eg: as far as I can see, cpuinfo, rdt caps
type Config struct {
	GenericConfig *GenericAPIConfig
	SwaggerConfig *swagger.Config
}

// TODO move this out of app server
type completedConfig struct {
	*Config
}

// TODO move this out of app server
type APIServer struct {
	HandlerContainer *restful.Container
	Server           *http.Server
}

// TODO move this out of app server
func NewAPIConfig() *GenericAPIConfig {
	return &GenericAPIConfig{
		APIServerServiceIP:   "localhost",
		APIServerServicePort: "8080",
		EnableUISupport:      true,
	}
}

func BuildServerConfig(s *options.ServerRunOptions) (*Config, error) {
	apiconfig := NewAPIConfig()

	// FIXME (cmd line options does not override the config file options)
	appconfig := appConf.NewConfig()
	if s.Addr == "" {
		s.Addr = appconfig.Def.Address
	}
	apiconfig.APIServerServiceIP = s.Addr
	if s.Port == "" {
		s.Port = strconv.FormatUint(uint64(appconfig.Def.Port), 10)
	}
	apiconfig.APIServerServicePort = s.Port
	apiconfig.Transport = appconfig.Db.Transport

	weburl := fmt.Sprintf("http://%s:%s", s.Addr, s.Port)
	swaggerconfig := swagger.Config{
		WebServicesUrl: weburl,
		ApiPath:        "/apidocs.json",
		// Optionally, specifiy where the UI is located
		SwaggerPath: "/apidocs/",
		// FIXME (eliqiao): this depends on https://github.com/swagger-api/swagger-ui.git need to copy dist from it
		SwaggerFilePath: "/usr/local/share/go/src/github.com/wordnik/swagger-ui/dist",
		ApiVersion:      "1.0",
	}

	config := &Config{
		GenericConfig: apiconfig,
		SwaggerConfig: &swaggerconfig,
	}
	return config, nil
}

func (c *Config) Complete() completedConfig {
	// TODO to complete config in this function

	// test init capabilities
	l3cat := capabilities.L3CAT{
		NumCLOS: 16,
		NumWays: 20,
		WaySize: 10000,
	}

	capabilities.Setup(nil, &l3cat, nil)

	return completedConfig{c}
}

func (c completedConfig) New() (*APIServer, error) {
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})

	cap := v1.CapabilitiesResource{}
	caches := v1.CachesResource{}
	policy := v1.PolicyResource{}

	// FIXME Add config support
	var db db.DB = new(db.BoltDB)
	db.Initialize(c.Config.GenericConfig.Transport)
	wls := v1.WorkLoadResource{Db: db}
	// Register controller to container
	cap.Register(wsContainer)
	caches.Register(wsContainer)
	policy.Register(wsContainer)
	wls.Register(wsContainer)

	// Install adds the SgaggerUI webservices
	c.Config.SwaggerConfig.WebServices = wsContainer.RegisteredWebServices()
	swagger.RegisterSwaggerService(*(c.Config.SwaggerConfig), wsContainer)

	// TODO error handle
	return &APIServer{
		HandlerContainer: wsContainer,
		Server:           &http.Server{Addr: c.Config.GenericConfig.APIServerServiceIP + ":" + c.Config.GenericConfig.APIServerServicePort, Handler: wsContainer},
	}, nil
}

// RunServer uses the provided options to run the apiserver.
func RunServer(s *options.ServerRunOptions) {
	// TODO handle err
	config, _ := BuildServerConfig(s)

	ser, _ := config.Complete().New()
	fmt.Println("started")
	log.Fatal(ser.Server.ListenAndServe())
}
