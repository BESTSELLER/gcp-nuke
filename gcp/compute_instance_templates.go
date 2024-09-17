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

// ComputeInstanceTemplates -
type ComputeInstanceTemplates struct {
	serviceClient *compute.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

// Name - Name of the resourceLister for ComputeInstanceTemplates
func (c *ComputeInstanceTemplates) Name() string {
	return "ComputeInstanceTemplates"
}

// ToSlice - Name of the resourceLister for ComputeInstanceTemplates
func (c *ComputeInstanceTemplates) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)
}

// Setup - populates the struct
func (c *ComputeInstanceTemplates) Setup(config config.Config) {
	c.base.config = config

	computeService, err := compute.NewService(Ctx, option.WithTokenSource(config.GCPToken))
	if err != nil {
		log.Fatalf("ComputeInstanceTemplates.Setup.NewService: %s", err)
	}
	computeResource := ComputeInstanceTemplates{
		serviceClient: computeService,
	}
	register(&computeResource)
}

// List - Returns a list of all ComputeInstanceTemplates
func (c *ComputeInstanceTemplates) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	instanceListCall := c.serviceClient.InstanceTemplates.List(c.base.config.Project)
	instanceList, err := instanceListCall.Do()
	if err != nil {
		log.Fatalf("ComputeInstanceTemplates.List: %s", err)
	}

	for _, instance := range instanceList.Items {
		instanceResource := DefaultResourceProperties{}
		c.resourceMap.Store(instance.Name, instanceResource)
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *ComputeInstanceTemplates) Dependencies() []string {
	a := ComputeInstanceGroupsRegion{}
	b := ComputeInstanceGroupsZone{}
	cl := ContainerGKEClusters{}
	return []string{a.Name(), b.Name(), cl.Name()}
}

// Remove -
func (c *ComputeInstanceTemplates) Remove() error {
	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		instanceID := key.(string)

		// Check if a resource is exclued from deletion
		if slices.Contains(c.base.config.Exclusions.ComputeInstanceTemplate, instanceID) {
			log.Printf("[Info] Excluded resource: %v (%v)", instanceID, c.Name())
			// This instanceID is excluded from deletion, returning
			return false
		}

		// Parallel instance deletion
		errs.Go(func() error {
			deleteCall := c.serviceClient.InstanceTemplates.Delete(c.base.config.Project, instanceID)
			operation, err := deleteCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being deleted %v [type: %v project: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, seconds)
				operationCall := c.serviceClient.GlobalOperations.Get(c.base.config.Project, operation.Name)
				checkOpp, err := operationCall.Do()
				if err != nil {
					return err
				}
				opStatus = checkOpp.Status

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource deletion timed out for %v [type: %v project: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(instanceID)

			log.Printf("[Info] Resource deleted %v [type: %v project: %v] (%v seconds)", instanceID, c.Name(), c.base.config.Project, seconds)
			return nil
		})
		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
