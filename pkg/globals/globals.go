package globals

const (
	DefaultStackConfigFileExtension = ".yaml"
	ConfigFileName                  = "atmos.yaml"
	SystemDirConfigFilePath         = "/usr/local/etc/atmos"
	WindowsAppDataEnvVar            = "LOCALAPPDATA"

	// GlobalOptionsFlag is a custom flag to specify helmfile `GLOBAL OPTIONS`
	// https://github.com/roboll/helmfile#cli-reference
	GlobalOptionsFlag = "--global-options"

	TerraformDirFlag = "--terraform-dir"
	HelmfileDirFlag  = "--helmfile-dir"
	ConfigDirFlag    = "--config-dir"
	StackDirFlag     = "--stacks-dir"
	BasePathFlag     = "--base-path"
	WorkflowDirFlag  = "--workflows-dir"

	DeployRunInitFlag           = "--deploy-run-init"
	AutoGenerateBackendFileFlag = "--auto-generate-backend-file"

	FromPlanFlag = "--from-plan"

	HelpFlag1 = "-h"
	HelpFlag2 = "--help"
)

var (
	LogVerbose = false
)
