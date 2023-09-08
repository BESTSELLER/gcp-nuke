package config

import (
	"context"
)

// Config -
type Config struct {
	Project  string
	Zones    []string
	Regions  []string
	Timeout  int
	PollTime int
	Context  context.Context
	DryRun   bool
  Exclusions Exclusions
}

type Exclusions struct {
	BigQuery []string
	ComputeDisk []string
	ComputeFirewall []string
	ComputeInstanceGroupsRegion []string
	ComputeInstanceGroupsZone []string
	ComputeInstanceTemplate []string
	ComputeInstance []string	
	ComputeNetworkPeering []string
	ComputeRegionAutoscaler []string
	ComputeRouter []string
	ComputeSubNetwork []string
	ComputeVPNGateway []string
	ComputeVPNTunnel []string
	ComputeZoneAutoscaler []string
	ContainerGKECluster []string
	GoogleComputeNetwork []string
	IAMServiceAccount []string
}
