package gcp

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/BESTSELLER/gcp-nuke/config"
	"github.com/BESTSELLER/gcp-nuke/helpers"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ComputeInstanceGroupsRegion -
type ComputeInstanceGroupsRegion struct {
	serviceClient *compute.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

// Name - Name of the resourceLister for ComputeInstanceGroupsRegion
func (c *ComputeInstanceGroupsRegion) Name() string {
	return "ComputeInstanceGroupsRegion"
}

// ToSlice - Name of the resourceLister for ComputeInstanceGroupsRegion
func (c *ComputeInstanceGroupsRegion) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)
}

// Setup - populates the struct
func (c *ComputeInstanceGroupsRegion) Setup(config config.Config) {
	c.base.config = config

	computeService, err := compute.NewService(Ctx, option.WithTokenSource(config.GCPToken))
	if err != nil {
		log.Fatalf("ComputeInstanceGroupsRegion.Setup.NewService: %s", err)
	}
	computeResource := ComputeInstanceGroupsRegion{
		serviceClient: computeService,
	}
	register(&computeResource)
}

// List - Returns a list of all ComputeInstanceGroupsRegion
func (c *ComputeInstanceGroupsRegion) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	for _, region := range c.base.config.Regions {
		instanceListCall := c.serviceClient.RegionInstanceGroupManagers.List(c.base.config.Project, region)
		instanceList, err := instanceListCall.Do()
		if err != nil {
			log.Fatalf("ComputeInstanceGroupsRegion.List: %s", err)
		}

		for _, instance := range instanceList.Items {
			instanceResource := DefaultResourceProperties{
				region: region,
			}
			c.resourceMap.Store(instance.Name, instanceResource)
		}
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *ComputeInstanceGroupsRegion) Dependencies() []string {
	a := ComputeRegionAutoScalers{}
	return []string{a.Name()}
}

// Remove -
func (c *ComputeInstanceGroupsRegion) Remove() error {
	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		instanceID := key.(string)
		region := value.(DefaultResourceProperties).region

		// Check if a resource is exclued from deletion
		if slices.Contains(c.base.config.Exclusions.ComputeInstanceGroupsRegion, instanceID) {
			log.Printf("[Info] Excluded resource: %v (%v)", instanceID, c.Name())
			// This instanceID is excluded from deletion, returning
			return false
		}

		// Parallel instance deletion
		errs.Go(func() error {
			deleteCall := c.serviceClient.RegionInstanceGroupManagers.Delete(c.base.config.Project, region, instanceID)
			operation, err := deleteCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being deleted %v [type: %v project: %v region: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, region, seconds)
				operationCall := c.serviceClient.RegionOperations.Get(c.base.config.Project, region, operation.Name)
				checkOpp, err := operationCall.Do()
				if err != nil {
					return err
				}
				opStatus = checkOpp.Status

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource deletion timed out for %v [type: %v project: %v region: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, region, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(instanceID)

			log.Printf("[Info] Resource deleted %v [type: %v project: %v region: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, region, seconds)
			return nil
		})
		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
