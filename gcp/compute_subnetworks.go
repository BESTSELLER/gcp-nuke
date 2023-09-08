package gcp

import (
	"fmt"
	"log"
	"sync"
	"time"

 "golang.org/x/exp/slices"
	"github.com/BESTSELLER/gcp-nuke/config"
	"github.com/BESTSELLER/gcp-nuke/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/compute/v1"
)

// ComputeSubnetworks -
type ComputeSubnetworks struct {
	serviceClient *compute.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	computeService, err := compute.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	computeResource := ComputeSubnetworks{
		serviceClient: computeService,
	}
	register(&computeResource)
}

// Name - Name of the resourceLister for ComputeSubnetworks
func (c *ComputeSubnetworks) Name() string {
	return "ComputeSubnetworks"
}

// ToSlice - Name of the resourceLister for ComputeSubnetworks
func (c *ComputeSubnetworks) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)

}

// Setup - populates the struct
func (c *ComputeSubnetworks) Setup(config config.Config) {
	c.base.config = config

}

// List - Returns a list of all ComputeSubnetworks
func (c *ComputeSubnetworks) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	for _, region := range c.base.config.Regions {
		subnetworkListCall := c.serviceClient.Subnetworks.List(c.base.config.Project, region)
		subnetworkList, err := subnetworkListCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		for _, subnetwork := range subnetworkList.Items {
			c.resourceMap.Store(subnetwork.Name, region)
		}
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *ComputeSubnetworks) Dependencies() []string {
	a := ComputeInstanceGroupsRegion{}
	b := ComputeInstanceGroupsZone{}
	cl := ContainerGKEClusters{}
	return []string{a.Name(), b.Name(), cl.Name()}
}

// Remove -
func (c *ComputeSubnetworks) Remove() error {

	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		subnetworkID := key.(string)
		region := value.(string)

    // Check if a resource is exclued from deletion
  	if slices.Contains(c.base.config.Exclusions.ComputeSubNetwork, subnetworkID) {
  		// This instanceID is excluded from deletion, returning
  		return false
		}

		// Parallel subnetwork deletion
		errs.Go(func() error {
			deleteCall := c.serviceClient.Subnetworks.Delete(c.base.config.Project, region, subnetworkID)
			operation, err := deleteCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being deleted %v [type: %v project: %v] (%v seconds)", subnetworkID, c.Name(), c.base.config.Project, seconds)

				operationCall := c.serviceClient.RegionOperations.Get(c.base.config.Project, region, operation.Name)
				checkOpp, err := operationCall.Do()
				if err != nil {
					return err
				}
				opStatus = checkOpp.Status

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource deletion timed out for %v [type: %v project: %v] (%v seconds)", subnetworkID, c.Name(), c.base.config.Project, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(subnetworkID)

			log.Printf("[Info] Resource deleted %v [type: %v project: %v] (%v seconds)", subnetworkID, c.Name(), c.base.config.Project, seconds)
			return nil
		})
		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
