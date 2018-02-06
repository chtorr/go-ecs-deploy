
go-ecs-deploy
=============

## Overview

A small library to take make deploying to AWS ECS from remote systems (like CI servers) a little easier.  It attempts to treat deployment like an "upsert" operation, where new services are created then deployed, and existing services are updated then re-deployed.  Deploying an existing service without any configuration changes will restart the service (useful if you are using a `latest` tag somewhere and just want to pull and restart).

The library supports two deployment types: `services` and `one-shots` (described below).

This repo includes a CLI wrapper that you can invoke directly (see below for usage), or you could include the library in another tool, like a deployment chat bot.


### Services
A service is specifically an [ECS Service](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html).  It is expected to start and remain running indefinitely.

### One-shots
A one-shot service is expected to run and exit.  It uses [ECS RunTask](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_run_task.html) under the hood.  One-shots are useful for running stand-alone database migration containers, or other ad-hoc one off tasks.

## CLI Tool

The included CLI tool takes raw ECS configuration and passes it through a single layer of templating with a global environment configuration object, to help deal with some of the natural config redundancy that tends to emerge.

The CLI tool consumes a single config file that contains a single raw ECS service and task config within (ECS tasks handle multi container pods on their own, this tool doesn't care).  Refer to the [ECS Task Definition docs](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) and [ECS Service docs](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service_definition_parameters.html) for specific details.

## Usage
AWS credentials must be available in the usual ways (metadata endpoint, credentials file, environment vars).

**Note:** the CLI tool does not yet support profiles, as the intended place to run is a CI job that most likely has env vars set.

Currently the tool assumes you have files laid out as detailed below in the [File Layout](#FileLayout) section.
```sh
$ go run src/main.go -h
  -debug
    	Print the the template, create the deployer, exit without deploying
  -dir string
    	The directory that holds config for the service
  -type string
    	The type of service to deploy {service | migration}
```


**Deploy a service:**
```sh
go run src/main.go -dir /path/to/environments/dev/service-a -type service
```
The above command will:
* load the `config.json` from the specified directory and the `environment.json` file from one level up.
* replace any template vars in the service config with values from the environment config.
* upsert the task and service in ECS
* block waiting for the service to be stable using the [ECS waiters](https://docs.aws.amazon.com/sdk-for-go/api/service/ecs/#ECS.WaitUntilServicesStable)




### File layout
This tool assumes your config files are organized as a directory for the environment with sub-directories for each service, with an `environment.json` at the root.  The environment config is a small list of values we'll replace in the service template.  

New values must be added to the Go struct.

```sh
dev/
    environment.json
    service-a/
        config.json
```


## Config Examples
The `environment.json` config looks like:
```json
{
    "Cluster": "some-cluster" ,
    "SchedulerIAMRoleArn": "arn",
    "TaskIAMRoleArn": "arn",
    "AWSLogsGroupName": "name",
    "AWSLogsRegion": "region"
}
```


An example service `config.json` looks like:
```json
{
    "service": {
        "cluster": "{{ .Cluster }}",
        "serviceName": "awesome-service",
        "loadBalancers": [
            {
                "targetGroupArn": "target group arn",
                "containerName": "awesome-service",
                "containerPort": 8080
            }
        ],
        "desiredCount": 2,
        "role": "{{ .SchedulerIAMRoleArn }}",
        "deploymentConfiguration": {
            "maximumPercent": 200,
            "minimumHealthyPercent": 100
        }
    },
    "task": {
        "family": "{{ .Cluster }}-awesome-service",
        "taskRoleArn": "{{ .TaskIAMRoleArn }}",
        "containerDefinitions": [
        {
            "name": "awesome-service",
            "image": "some-repo/awesome-service:tag",
            "essential": true,
            "memoryReservation": 250,
            "portMappings": [
            {
                "containerPort": 8080,
                "hostPort": 0
            }
            ],
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "{{ .AWSLogsGroupName }}",
                    "awslogs-region": "{{ .AWSLogsRegion }}",
                    "awslogs-stream-prefix": "awesome-service"
                }
            },
            "environment": [
            {
                "name": "SOME_VAR",
                "value": "SOME_KEY"
            }
            ]
        }
        ]
    }
}
```