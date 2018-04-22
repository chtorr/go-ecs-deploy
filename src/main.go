package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/chtorr/go-ecs-deploy/src/ecsdeploy"
)

// cmd deploy -configDir string []-timeout int]

const (
	defaultTimeout            = 360 * time.Second
	configFilename            = "config.json"
	environmentConfigFilename = "environment.json"
	ServiceTypeService        = "service"
	ServiceTypeMigration      = "migration"
)

// type EnvironmentConfig struct {
// 	Cluster             string
// 	SchedulerIAMRoleArn string
// 	TaskIAMRoleArn      string
// 	AWSLogsGroupName    string
// 	AWSLogsRegion       string
// }

type Config struct {
	Service ecs.CreateServiceInput
	Task    ecs.RegisterTaskDefinitionInput
}

func main() {
	config := flag.String("config", "", "The config file to load")
	serviceType := flag.String("type", "", "The type of service to deploy {service | migration}")
	debug := flag.Bool("debug", false, "Print the the template, create the deployer, exit without deploying")
	flag.Parse()

	if *config == "" {
		log.Println("config cannot be blank")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *serviceType == "" {
		log.Println("type cannot be blank")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *serviceType != ServiceTypeService && *serviceType != ServiceTypeService {
		log.Printf("Invalid service type: (%s)", *serviceType)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// clean up the passed in config dir path
	// configDir := filepath.Clean(*dir)
	// log.Printf("config dir: %s", configDir)

	// load the environment config
	// envConfig, err := loadEnvironmentConfig(configDir, environmentConfigFilename)
	// if err != nil {
	// 	log.Fatalf("Failed loading environment config: %v", err)
	// }

	// if envConfig.Cluster == "" {
	// 	log.Fatalf("Cluster must be provided in environment config.")
	// }

	// load service/task config
	c, err := loadConfig(*config)
	if err != nil {
		log.Fatalf("Failed loading service config: %v", err)
	}

	// Create the deployer
	deployer := ecsdeploy.NewECSDeployer(*c.Service.Cluster)

	if *debug {
		log.Println(config)
		os.Exit(0)
	}

	// Deploy
	var deployErr error

	if *serviceType == ServiceTypeService {
		log.Printf("Deploying service...")
		deployErr = deployer.DeployService(&c.Task, &c.Service)
	} else {
		log.Printf("Deploying migration...")
		deployErr = deployer.DeployOneshot(&c.Task)
	}

	if deployErr != nil {
		log.Fatalf("Deployment failed: %v", deployErr)
	}

	// log.Printf("Deploy success")
	os.Exit(0)
}

// func loadEnvironmentConfig(configDir, filename string) (env EnvironmentConfig, err error) {
// 	baseDir := filepath.Dir(configDir)
// 	fname := filepath.Join(baseDir, filename)
// 	f, err := ioutil.ReadFile(fname)
// 	if err != nil {
// 		return
// 	}

// 	err = json.Unmarshal(f, &env)
// 	if err != nil {
// 		return
// 	}

// 	return env, nil
// }

func loadConfig(filename string) (config Config, err error) {
	// fname := filepath.Join(configDir, filename)
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	// t, err := template.New("config").Parse(string(f))
	// if err != nil {
	// 	return
	// }

	// var buf bytes.Buffer
	// err = t.Execute(&buf, envConfig)
	// if err != nil {
	// 	return
	// }

	var c Config
	err = json.Unmarshal(f, &c)
	if err != nil {
		return
	}

	return c, nil
}
