package gcp

import (
	"context"
	"log"
	"strings"

	"github.com/BESTSELLER/gcp-nuke/config"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ResourceBase -
type ResourceBase struct {
	config config.Config
}

// DefaultResourceProperties -
type DefaultResourceProperties struct {
	project string
	zone    string
	region  string
}

// Resource -
type Resource interface {
	Name() string
	ToSlice() []string
	Setup(config config.Config)
	List(useCache bool) []string
	Dependencies() []string
	Remove() error
}

// Ctx = context
var (
	Ctx         = context.Background()
	resourceMap = make(map[string]Resource)
)

func register(resource Resource) {
	_, exists := resourceMap[resource.Name()]
	if exists {
		log.Fatalf("a resource with the name %s already exists", resource.Name())
	}
	resourceMap[resource.Name()] = resource
}

// GetResourceMap -
func GetResourceMap(config config.Config) map[string]Resource {
	for _, resource := range resourceMap {
		resource.Setup(config)
	}

	return resourceMap
}

// GetZones -
func GetZones(defaultContext context.Context, project string) []string {
	log.Println("[Info] Retrieving zones for project:", project)
	client, err := google.DefaultClient(defaultContext, compute.ComputeScope)
	if err != nil {
		log.Fatalf("GetZones.DefaultClient: %s", err)
	}
	serviceClient, err := compute.New(client)
	if err != nil {
		log.Fatalf("GetZones.serviceClient: %s", err)
	}
	zoneListCall := serviceClient.Zones.List(project)
	zoneList, err := zoneListCall.Do()
	if err != nil {
		log.Fatalf("GetZones.Zones.List: %s", err)
	}

	zoneStringSlice := []string{}
	for _, zone := range zoneList.Items {
		zoneNameSplit := strings.Split(zone.Name, "/")
		zoneStringSlice = append(zoneStringSlice, zoneNameSplit[len(zoneNameSplit)-1])
	}
	return zoneStringSlice
}

func AddZonesToConfig(defaultContext context.Context, project string, config config.Config) {
	computeService, err := compute.NewService(defaultContext, option.WithTokenSource(config.GCPToken))
	if err != nil {
		log.Fatalf("AddZonesToConfig.NewService: %s", err)
	}
	zones, err := computeService.Zones.List(project).Do()
	if err != nil {
		log.Fatalf("AddZonesToConfig.List: %s", err)
	}
	config.Zones = []string{}
	for _, zone := range zones.Items {
		zoneNameSplit := strings.Split(zone.Name, "/")
		config.Zones = append(config.Zones, zoneNameSplit[len(zoneNameSplit)-1])
	}
}

// GetRegions -
func GetRegions(defaultContext context.Context, project string) []string {
	log.Println("[Info] Retrieving regions for project:", project)
	client, err := google.DefaultClient(defaultContext, compute.ComputeScope)
	if err != nil {
		log.Fatalf("GetRegions.DefaultClient: %s", err)
	}
	serviceClient, err := compute.New(client)
	if err != nil {
		log.Fatalf("GetRegions.serviceClient: %s", err)
	}
	regionListCall := serviceClient.Regions.List(project)
	regionList, err := regionListCall.Do()
	if err != nil {
		log.Fatalf("GetRegions.Regions.List: %s", err)
	}

	regionStringSlice := []string{}
	for _, region := range regionList.Items {
		regionNameSplit := strings.Split(region.Name, "/")
		regionStringSlice = append(regionStringSlice, regionNameSplit[len(regionNameSplit)-1])
	}
	return regionStringSlice
}

func AddRegionsToConfig(defaultContext context.Context, project string, config config.Config) {
	computeService, err := compute.NewService(defaultContext, option.WithTokenSource(config.GCPToken))
	if err != nil {
		log.Fatalf("AddRegionsToConfig.NewService: %s", err)
	}
	regions, err := computeService.Regions.List(project).Do()
	if err != nil {
		log.Fatalf("AddRegionsToConfig.List: %s", err)
	}
	config.Regions = []string{}
	for _, region := range regions.Items {
		regionNameSplit := strings.Split(region.Name, "/")
		config.Regions = append(config.Regions, regionNameSplit[len(regionNameSplit)-1])
	}
}

func extractGKESelfLink(input string) string {
	var selfLinkSlice []string
	var startAppend bool
	for _, word := range strings.Split(input, "/") {
		if word == "projects" {
			startAppend = true
		}
		if word == "zones" {
			word = "locations"
		}
		if startAppend {
			selfLinkSlice = append(selfLinkSlice, word)
		}
	}
	return strings.Join(selfLinkSlice, "/")
}
