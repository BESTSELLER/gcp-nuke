package gcp

import (
	"log"
	"strings"
	"sync"

	"github.com/BESTSELLER/gcp-nuke/config"
	"github.com/BESTSELLER/gcp-nuke/helpers"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type IAMServiceAccount struct {
	serviceClient *iam.Service
	base          ResourceBase
	resourceMap   syncmap.Map
	TopicIDs      []string
}

func (c *IAMServiceAccount) Name() string {
	return "IAMServiceAccount"
}

func (c *IAMServiceAccount) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)
}

func (c *IAMServiceAccount) Setup(config config.Config) {
	c.base.config = config

	iamService, err := iam.NewService(Ctx, option.WithTokenSource(config.GCPToken))
	if err != nil {
		log.Fatalf("IAMServiceAccount.Setup.NewService: %s", err)
	}

	iamResource := IAMServiceAccount{
		serviceClient: iamService,
	}
	register(&iamResource)
}

func (c *IAMServiceAccount) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	// Refresh resource map
	c.resourceMap = sync.Map{}

	serviceAccountList, err := c.serviceClient.Projects.ServiceAccounts.List("projects/" + c.base.config.Project).Do()
	if err != nil {
		log.Fatalf("IAMServiceAccount.List: %s", err)
	}

	for _, serviceAccount := range serviceAccountList.Accounts {
		// Will not list / delete default service accounts
		if strings.Contains(serviceAccount.Email, c.base.config.Project) {
			c.resourceMap.Store(serviceAccount.DisplayName, serviceAccount.Email)
		}
	}

	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *IAMServiceAccount) Dependencies() []string {
	return []string{}
}

func (c *IAMServiceAccount) Remove() error {
	// Removal logic
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		emailAddress := value.(string)

		// Check if a resource is exclued from deletion
		if slices.Contains(c.base.config.Exclusions.IAMServiceAccount, emailAddress) {
			log.Printf("[Info] Excluded resource: %v (%v)", emailAddress, c.Name())
			// This instanceID is excluded from deletion, returning
			return false
		}

		// Parallel instance deletion
		errs.Go(func() error {
			_, err := c.serviceClient.Projects.ServiceAccounts.Delete("projects/" + c.base.config.Project + "/serviceAccounts/" + emailAddress).Do()
			if err != nil {
				return err
			}

			seconds := 0

			log.Printf("[Info] Resource deleted %v [type: %v project: %v] (%v seconds)", emailAddress, c.Name(), c.base.config.Project, seconds)
			return nil
		})
		return true
	})
	// Wait for all deletions to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
