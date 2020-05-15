package api

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
)

func swagger(c *restful.Container, address string) {
	config := restfulspec.Config{
		WebServices:                   c.RegisteredWebServices(),
		WebServicesURL:                fmt.Sprintf(":%s", address),
		APIPath:                       "/api/docs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	c.Handle("/api/docs/", http.StripPrefix("/api/docs/", http.FileServer(http.Dir("dist"))))
	c.Add(restfulspec.NewOpenAPIService(config))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Kubernetes Transform APIServer",
			Description: "Resource for managing Kubernetes Transform API",
			Version:     "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
		Name:        "apis",
		Description: "Managing API"}}}
}
