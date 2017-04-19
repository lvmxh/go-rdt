package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-swagger12"
	"openstackcore-rdtagent/api/v1"
	"openstackcore-rdtagent/pkg/model/capabilities"
	"openstackcore-rdtagent/util/options"
)

// TODO move this out of app server
type GenericAPIConfig struct {
	//APIServerServiceIP   net.IP
	APIServerServiceIP   string
	APIServerServicePort string
	EnableUISupport      bool
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

	if s.Addr != "" {
		apiconfig.APIServerServiceIP = s.Addr
	}
	if s.Port != "" {
		apiconfig.APIServerServicePort = s.Port
	}

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

	cpuinfo := v1.CpuinfoResource{}
	caches := v1.CachesResource{}
	// Register controller to container
	cpuinfo.Register(wsContainer)
	caches.Register(wsContainer)

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
