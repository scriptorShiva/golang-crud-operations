package config

import (
	"log" // Log fatal errors
	"os"  // Read env vars, check file existence

	"flag" // Read CLI flags like --config=...

	"github.com/ilyakaznacheev/cleanenv" // Load config from YAML + env vars
)

type HTTPServer struct {
	Host string
	Port string
}

// This is your application config contract.
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true" ` // load from .env key in yaml or from ENV , app fail to start if missing
	StoragePath string `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

// in go convention : load return (cfg , err), mustload : exits the program on failure
func MustLoad() *Config {
	var configPath string

	// “Hey OS, do you have a variable called CONFIG_PATH? If yes, give me its value.” 
	configPath = os.Getenv("CONFIG_PATH")

	// fallback to CLI flag - If env not set read --config flag , if misses crash immediately
	if configPath == "" {
		
		// flags are the ones we pass when we run our program go run ./cmd/crud-operations/main.go --config=config/local.yaml
		flags := flag.String("config" , "" , "path to the configuration file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	// does the config file actually exist . If not crash it.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("Config file does not exist")
	}

	var cfg Config

	// load env from clearconfig. what is does - read yaml from file , read env vars , merge them , applies rules env-required "true", maps values to struct 
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Unable to read config: %v", err.Error())
	}

	// why pointer : avoid copying struct , shared real-only config, standard go practice
	return &cfg
}


/*
OS sets env / flags
↓
MustLoad()
↓
Validate config
↓
Read config
↓
Return Config
↓
App starts

*/