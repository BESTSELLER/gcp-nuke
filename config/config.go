package config

import (
	"context"

	"golang.org/x/oauth2"
)

// Config -
type Config struct {
	Project    string
	Zones      []string
	Regions    []string
	Timeout    int
	PollTime   int
	Context    context.Context
	DryRun     bool
	Exclusions Exclusions
	GCPToken   oauth2.TokenSource
}

type Exclusions struct {
	BigQuery                    []string `json:"bigquery"`
	ComputeDisk                 []string `json:"compute_disk"`
	ComputeFirewall             []string `json:"compute_firewall"`
	ComputeInstanceGroupsRegion []string `json:"compute_instance_groups_region"`
	ComputeInstanceGroupsZone   []string `json:"compute_instance_groups_zone"`
	ComputeInstanceTemplate     []string `json:"compute_instance_template"`
	ComputeInstance             []string `json:"compute_instance"`
	ComputeNetworkPeering       []string `json:"compute_network_peering"`
	ComputeRegionAutoscaler     []string `json:"compute_region_autoscaler"`
	ComputeRouter               []string `json:"compute_router"`
	ComputeSubNetwork           []string `json:"compute_subnetwork"`
	ComputeVPNGateway           []string `json:"compute_vpn_gateway"`
	ComputeVPNTunnel            []string `json:"compute_vpn_tunnel"`
	ComputeZoneAutoscaler       []string `json:"compute_zone_autoscaler"`
	ContainerGKECluster         []string `json:"container_gke_cluster"`
	GoogleComputeNetwork        []string `json:"google_compute_network"`
	IAMServiceAccount           []string `json:"iam_service_account"`
}

func ConvertStringToTokenSource(token string) oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
}
