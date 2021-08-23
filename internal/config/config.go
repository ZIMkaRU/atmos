package config

import (
	g "atmos/internal/globals"
	u "atmos/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/bmatcuk/doublestar"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	configFileName          = "atmos.yaml"
	systemDirConfigFilePath = "/usr/local/etc/atmos"
	windowsAppDataEnvVar    = "LOCALAPPDATA"
)

type Configuration struct {
	StackNamePattern         string   `mapstructure:"StackNamePattern"`
	StackDirs                []string `mapstructure:"StackDirs"`
	StackDirsAbsolutePaths   []string
	TerraformDir             string `mapstructure:"TerraformDir"`
	TerraformDirAbsolutePath string
	StackConfigFiles         []string
}

var (
	// Default values
	defaultConfig = map[string]interface{}{
		// Default paths to stack configs
		"StackDirs": []interface{}{
			"./stacks/*",
		},
		// Default path to terraform components
		"TerraformDir": "./components/terraform",
		// Logical stack name pattern
		"StackNamePattern": "environment-stage",
	}

	// Config is the CLI configuration structure
	Config Configuration
)

// https://dev.to/techschoolguru/load-config-from-file-environment-variables-in-golang-with-viper-2j2d
// https://medium.com/@bnprashanth256/reading-configuration-files-and-environment-variables-in-go-golang-c2607f912b63
func InitConfig() error {
	// Config is loaded from these locations (from lower to higher priority):
	// /usr/local/etc/atmos
	// ~/.atmos
	// from the current directory
	// from ENV vars
	// from command-line arguments

	fmt.Println(strings.Repeat("-", 120))
	fmt.Println("Processing and merging configurations in the following order: system dir, home dir, current dir, ENV vars")

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetTypeByDefaultValue(true)

	// Add default config
	err := v.MergeConfigMap(defaultConfig)
	if err != nil {
		return err
	}

	// Process config in system folder
	configFilePath1 := ""

	// https://pureinfotech.com/list-environment-variables-windows-10/
	// https://docs.microsoft.com/en-us/windows/deployment/usmt/usmt-recognized-environment-variables
	// https://softwareengineering.stackexchange.com/questions/299869/where-is-the-appropriate-place-to-put-application-configuration-files-for-each-p
	// https://stackoverflow.com/questions/37946282/why-does-appdata-in-windows-7-seemingly-points-to-wrong-folder
	if runtime.GOOS == "windows" {
		appDataDir := os.Getenv(windowsAppDataEnvVar)
		if len(appDataDir) > 0 {
			configFilePath1 = appDataDir
		}
	} else {
		configFilePath1 = systemDirConfigFilePath
	}

	if len(configFilePath1) > 0 {
		configFile1 := path.Join(configFilePath1, configFileName)
		err = processConfigFile(configFile1, v)
		if err != nil {
			return err
		}
	}

	// Process config in user's HOME dir
	configFilePath2, err := homedir.Dir()
	if err != nil {
		return err
	}
	configFile2 := path.Join(configFilePath2, ".atmos", configFileName)
	err = processConfigFile(configFile2, v)
	if err != nil {
		return err
	}

	// Process config in the current dir
	configFilePath3, err := os.Getwd()
	if err != nil {
		return err
	}
	configFile3 := path.Join(configFilePath3, configFileName)
	err = processConfigFile(configFile3, v)
	if err != nil {
		return err
	}

	// https://gist.github.com/chazcheadle/45bf85b793dea2b71bd05ebaa3c28644
	// https://sagikazarmark.hu/blog/decoding-custom-formats-with-viper/
	err = v.Unmarshal(&Config)
	if err != nil {
		return err
	}

	// Process ENV vars
	processEnvVars()

	// Check config
	err = checkConfig()
	if err != nil {
		return err
	}

	// Convert all stack dirs to absolute paths
	absPaths, err := u.ConvertPathsToAbsolutePaths(Config.StackDirs)
	if err != nil {
		return err
	}
	Config.StackDirsAbsolutePaths = absPaths

	// Convert terraform dir to absolute path
	terraformDirAbsPath, err := filepath.Abs(Config.TerraformDir)
	if err != nil {
		return err
	}
	Config.TerraformDirAbsolutePath = terraformDirAbsPath

	// Find all stack config files in the provided paths
	stackConfigFiles, err := findAllStackConfigsInPaths(absPaths)
	if err != nil {
		return err
	}
	if len(stackConfigFiles) < 1 {
		j, _ := json.MarshalIndent(absPaths, "", strings.Repeat(" ", 2))
		if err != nil {
			return err
		}
		errorMessage := fmt.Sprintf("No config files found in any of the provided path globs:\n%s\n", j)
		return errors.New(errorMessage)
	}
	Config.StackConfigFiles = stackConfigFiles

	fmt.Println("Final CLI configuration:")
	j, _ := json.MarshalIndent(&Config, "", strings.Repeat(" ", 2))
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)

	return nil
}

// https://github.com/NCAR/go-figure
// https://github.com/spf13/viper/issues/181
// https://medium.com/@bnprashanth256/reading-configuration-files-and-environment-variables-in-go-golang-c2607f912b63
func processConfigFile(path string, v *viper.Viper) error {
	if !u.FileExists(path) {
		fmt.Println("No config found at " + path)
		return nil
	}

	fmt.Println("Found config at " + path)

	reader, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			fmt.Println("Error closing file " + path + ". " + err.Error())
		}
	}(reader)

	err = v.MergeConfig(reader)
	if err != nil {
		return err
	}

	fmt.Println("Processed config at " + path)

	return nil
}

func processEnvVars() {
	stackDirs := os.Getenv("ATMOS_STACK_DIRS")
	if len(stackDirs) > 0 {
		fmt.Println("Found ENV var 'ATMOS_STACK_DIRS'")
		Config.StackDirs = strings.Split(stackDirs, ",")
	}

	terraformDir := os.Getenv("ATMOS_TERRAFORM_DIR")
	if len(terraformDir) > 0 {
		fmt.Println("Found ENV var 'ATMOS_TERRAFORM_DIR'")
		Config.TerraformDir = terraformDir
	}

	stackNamePattern := os.Getenv("ATMOS_STACK_NAME_PATTERN")
	if len(stackNamePattern) > 0 {
		fmt.Println("Found ENV var 'ATMOS_STACK_NAME_PATTERN'")
		Config.StackNamePattern = stackNamePattern
	}
}

func checkConfig() error {
	if len(Config.StackDirs) < 1 {
		return errors.New("At least one path to stack config must be provided in 'StackDirs' or 'ATMOS_STACK_DIRS' ENV variable")
	}
	if len(Config.TerraformDir) < 1 {
		return errors.New("Terraform dir must be provided in 'TerraformDir' or 'ATMOS_TERRAFORM_DIR' ENV variable")
	}
	if len(Config.StackNamePattern) < 1 {
		return errors.New("Stack name pattern must be provided in 'StackNamePattern' or 'ATMOS_STACK_NAME_PATTERN' ENV variable")
	}

	return nil
}

// findAllStackConfigsInPaths finds all stack config files in the paths specified by globs
func findAllStackConfigsInPaths(pathGlobs []string) ([]string, error) {
	res := []string{}

	for _, p := range pathGlobs {
		pathWithExt := p

		ext := filepath.Ext(p)
		if ext == "" {
			ext = g.DefaultStackConfigFileExtension
			pathWithExt = p + ext
		}

		// Find all matches in the glob
		matches, err := doublestar.Glob(pathWithExt)
		if err != nil {
			return nil, err
		}

		if matches != nil && len(matches) > 0 {
			res = append(res, matches...)
		}

	}

	return res, nil
}
