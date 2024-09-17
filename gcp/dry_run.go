package gcp

import (
	"log"

	"github.com/BESTSELLER/gcp-nuke/config"
)

func parallelDryRun(resource Resource, config config.Config) {
	resourceList := resource.List(false)
	if len(resourceList) == 0 {
		log.Printf("[Dryrun] [Skip] Resource type %v has nothing to destroy [project: %v]", resource.Name(), config.Project)
		return
	}
	log.Printf("[Dryrun] Resource type %v with resources %v would be destroyed [project: %v]", resource.Name(), resourceList, config.Project)
}
