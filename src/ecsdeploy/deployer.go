package ecsdeploy

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECSDeployer struct {
	client  *ecs.ECS
	cluster string
}

func makeStrPtr(s string) *string {
	return &s
}

func makeInt64Ptr(i int64) *int64 {
	return &i
}

func NewECSDeployer(cluster string) *ECSDeployer {
	return &ECSDeployer{
		client:  ecs.New(session.New()),
		cluster: cluster,
	}
}

func (e *ECSDeployer) DeployService(task *ecs.RegisterTaskDefinitionInput, service *ecs.CreateServiceInput) error {
	return e.deployService(task, service)
}

func (e *ECSDeployer) DeployOneshot(task *ecs.RegisterTaskDefinitionInput) error {
	return e.deployOneshot(task)
}

// Takes config for a task definiteion to be registered and returns a description of a registered task
func (e *ECSDeployer) registerTask(taskDef *ecs.RegisterTaskDefinitionInput) (*ecs.TaskDefinition, error) {
	// - register task definition, creating a new revision every time
	task, err := e.client.RegisterTaskDefinition(taskDef)
	if err != nil {
		return nil, err
	}

	return task.TaskDefinition, nil
}

// Takes a registered task definition and service config and updates an existing service
func (e *ECSDeployer) updateService(task *ecs.TaskDefinition, service *ecs.CreateServiceInput) (*ecs.Service, error) {
	s := &ecs.UpdateServiceInput{
		Cluster:                 service.Cluster,
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		Service:                 service.ServiceName,
		TaskDefinition:          task.TaskDefinitionArn,
	}

	so, err := e.client.UpdateService(s)
	if err != nil {
		return nil, err
	}

	return so.Service, nil
}

// Takes a registered task def and a service config and creates a new service
func (e *ECSDeployer) createService(task *ecs.TaskDefinition, service *ecs.CreateServiceInput) (*ecs.Service, error) {

	so, err := e.client.CreateService(&ecs.CreateServiceInput{
		ClientToken:             makeStrPtr(strconv.Itoa(int(time.Now().Unix()))),
		Cluster:                 service.Cluster,
		DeploymentConfiguration: service.DeploymentConfiguration,
		DesiredCount:            service.DesiredCount,
		LoadBalancers:           service.LoadBalancers,
		PlacementConstraints:    service.PlacementConstraints,
		TaskDefinition:          task.TaskDefinitionArn,
		Role:                    service.Role,
		ServiceName:             service.ServiceName,
	})
	if err != nil {
		return nil, err
	}
	return so.Service, err
}
