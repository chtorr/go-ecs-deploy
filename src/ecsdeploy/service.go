package ecsdeploy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ecs"
)

func (e *ECSDeployer) deployService(taskDef *ecs.RegisterTaskDefinitionInput, service *ecs.CreateServiceInput) error {

	task, err := e.registerTask(taskDef)
	if err != nil {
		return err
	}

	// We must check if the service already exists.
	// Depending on current state we will either Create a new service, or Update an existing service.
	// Currently there is no "upsert" like functionality as part of the ECS SDK.
	svcs, err := e.client.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: service.Cluster,
		Services: []*string{
			service.ServiceName,
		},
	})
	if err != nil {
		return err
	}

	action := "create"
	var s *ecs.Service

	// AWS enforces unique service names per cluster so this is unlikely,
	// but let's guard against multiple services in the response just in case.
	if len(svcs.Services) > 1 {
		return fmt.Errorf("More than one running service matches the provided name: %v", service.ServiceName)
	}

	if len(svcs.Services) == 1 && isServiceActive(svcs.Services[0]) {
		action = "update"
	}

	switch action {
	case "create":
		log.Printf("Creating service...")
		s, err = e.createService(task, service)
		if err != nil {
			return err
		}
	case "update":
		log.Printf("Updating service...")
		s, err = e.updateService(task, service)
		if err != nil {
			return err
		}
	}

	log.Printf("waiting for stable service state...")
	err = e.client.WaitUntilServicesStable(&ecs.DescribeServicesInput{
		Cluster: service.Cluster,
		Services: []*string{
			s.ServiceName,
		},
	})

	return err
}

func isServiceActive(s *ecs.Service) bool {
	if s.Status != nil {
		if *s.Status == "ACTIVE" {
			return true
		}
	}
	return false
}
