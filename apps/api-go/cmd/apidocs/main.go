package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"smartdelivery/apps/api-go/internal/httpserver"
	"smartdelivery/apps/api-go/internal/model"
	"smartdelivery/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type openAPISpec struct {
	OpenAPI    string                         `json:"openapi"`
	Info       openAPIInfo                    `json:"info"`
	Paths      map[string]map[string]pathItem `json:"paths"`
	XGenerated bool                           `json:"x-generated"`
}

type openAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type pathItem struct {
	Summary     string                 `json:"summary"`
	OperationID string                 `json:"operationId"`
	Responses   map[string]apiResponse `json:"responses"`
}

type apiResponse struct {
	Description string `json:"description"`
}

func main() {
	output := flag.String("out", "../../docs/generated/openapi.json", "output path for generated OpenAPI skeleton")
	flag.Parse()

	spec, err := generateSpec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate api docs: %v\n", err)
		os.Exit(1)
	}

	if err := writeSpec(*output, spec); err != nil {
		fmt.Fprintf(os.Stderr, "write api docs: %v\n", err)
		os.Exit(1)
	}
}

func generateSpec() (openAPISpec, error) {
	routes, err := collectRoutes()
	if err != nil {
		return openAPISpec{}, err
	}

	spec := openAPISpec{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:   "Smart Delivery API",
			Version: "0.1.0",
		},
		Paths:      make(map[string]map[string]pathItem),
		XGenerated: true,
	}

	for _, route := range routes {
		path := openAPIPath(route.Path)
		method := strings.ToLower(route.Method)
		if spec.Paths[path] == nil {
			spec.Paths[path] = make(map[string]pathItem)
		}
		spec.Paths[path][method] = describeRoute(route.Method, path)
	}

	return spec, nil
}

func collectRoutes() ([]gin.RouteInfo, error) {
	gin.SetMode(gin.ReleaseMode)

	handler := httpserver.NewHandler(httpserver.Dependencies{
		DeliveryRules: docsDeliveryRuleService{},
	})

	engine, ok := handler.(*gin.Engine)
	if !ok {
		return nil, errors.New("http handler is not a gin engine")
	}

	routes := engine.Routes()
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	return routes, nil
}

func openAPIPath(path string) string {
	parts := strings.Split(path, "/")
	for index, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[index] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func describeRoute(method string, path string) pathItem {
	key := method + " " + path
	summary := map[string]string{
		"GET /healthz":                              "Health check",
		"POST /shops/{shop}/delivery-rules":         "Create delivery rule",
		"GET /shops/{shop}/delivery-rules":          "List delivery rules",
		"PATCH /shops/{shop}/delivery-rules/{id}":   "Update delivery rule status",
		"POST /shops/{shop}/delivery-rules/imports": "Validate delivery rule import",
	}[key]
	if summary == "" {
		summary = method + " " + path
	}

	return pathItem{
		Summary:     summary,
		OperationID: operationID(method, path),
		Responses: map[string]apiResponse{
			defaultSuccessStatus(method, path): {Description: "Successful response"},
			"400":                              {Description: "Bad request"},
			"500":                              {Description: "Internal server error"},
		},
	}
}

func operationID(method string, path string) string {
	path = strings.Trim(path, "/")
	replacer := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_")
	name := replacer.Replace(path)
	if name == "" {
		return strings.ToLower(method)
	}
	return strings.ToLower(method) + "_" + name
}

func defaultSuccessStatus(method string, path string) string {
	if method == http.MethodPost && path == "/shops/{shop}/delivery-rules" {
		return "201"
	}
	return "200"
}

func writeSpec(output string, spec openAPISpec) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(spec); err != nil {
		return fmt.Errorf("encode openapi spec: %w", err)
	}

	return nil
}

type docsDeliveryRuleService struct{}

func (docsDeliveryRuleService) CreateRule(context.Context, service.CreateRuleCommand) (*model.DeliveryRule, error) {
	return nil, nil
}

func (docsDeliveryRuleService) ListRules(context.Context, service.ListRulesCommand) ([]*model.DeliveryRule, error) {
	return nil, nil
}

func (docsDeliveryRuleService) UpdateRuleStatus(context.Context, service.UpdateRuleStatusCommand) (*model.DeliveryRule, error) {
	return nil, nil
}

func (docsDeliveryRuleService) ValidateRuleImport(context.Context, service.ValidateRuleImportCommand) (service.RuleImportValidationSummary, error) {
	return service.RuleImportValidationSummary{}, nil
}
