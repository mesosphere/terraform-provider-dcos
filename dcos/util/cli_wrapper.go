package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/dcos/client-go/dcos"
)

var rxConfigArg = regexp.MustCompile(`([><] *)?%CONFIG%`)
var rxIDArg = regexp.MustCompile(`%NAME%`)

/**
 * The CLI wrapper provides some cached information about the CLI tools
 * used by the client
 */
type CliWrapper struct {
	sandbox     string
	client      *dcos.APIClient
	cliBinary   string
	dcosVersion string
	instances   map[string]*CliWrapperPackage
}

func CreateCliWrapper(sandbox string, client *dcos.APIClient, cli_version string) (*CliWrapper, error) {
	absSandbox, err := filepath.Abs(sandbox)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve absolute sandbox path: %s", err.Error())
	}

	return &CliWrapper{
		sandbox:     absSandbox,
		client:      client,
		cliBinary:   "",
		dcosVersion: cli_version,
		instances:   make(map[string]*CliWrapperPackage),
	}, nil
}

type PackageListCommand struct {
	Name string `json:"name"`
}

type PackageListResonse struct {
	PackageName string              `json:"name"`
	Command     *PackageListCommand `json:"command"`
}

type CliWrapperConfigParseError struct {
	Message string
}

func (e *CliWrapperConfigParseError) Error() string {
	return e.Message
}

type CliWrapperCommandNotFound struct {
	Message string
}

func (e *CliWrapperCommandNotFound) Error() string {
	return e.Message
}

/**
 * The CLI wrapper package provides the interface to package-specific commands
 */
type CliWrapperPackage struct {
	parent         *CliWrapper
	PackageName    string
	PackageCommand string
}

type CliWrapperPackageWithConfig struct {
	Package    *CliWrapperPackage
	ID         string
	Config     map[string]interface{}
	LastOutput string
}

func (w *CliWrapper) cliExec(args ...string) *exec.Cmd {
	log.Printf("Executing: %s %s", w.cliBinary, strings.Join(args, " "))
	cmd := exec.Command(w.cliBinary, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DCOS_DIR=%s/dcos", w.sandbox),
	)

	log.Printf("With env: %s", cmd.Env)
	return cmd
}

func (w *CliWrapper) shellExec(script string) *exec.Cmd {
	log.Printf("Executing shell script: %s", script)
	// TODO: Use OS shell
	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DCOS_DIR=%s/dcos", w.sandbox),
		fmt.Sprintf("DCOS_CLI=%s", w.cliBinary),
	)

	log.Printf("With env: %s", cmd.Env)
	return cmd
}

/**
 * Get a configuration property
 */
func (w *CliWrapper) GetConfig(name string) (string, error) {
	ret, err := w.cliExec("config", "get", name).Output()
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

/**
 * Update a configuration property
 */
func (w *CliWrapper) SetConfig(name string, value string) error {
	_, err := w.cliExec("config", "set", name, value).Output()
	if err != nil {
		return err
	}

	return nil
}

/**
 * Ensure that the cli for the correct DC/OS version exists in the sandbox
 */
func (w *CliWrapper) Prepare() error {
	log.Printf("Checking sandbox: %s", w.sandbox)

	// Ensure sandbox structure
	os.MkdirAll(fmt.Sprintf("%s/bin", w.sandbox), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/dcos", w.sandbox), os.ModePerm)

	// If the DC/OS version is not defined, pull the version now
	if w.dcosVersion == "" {
		log.Printf("dcos version is not cached, updating now")
		vers, err := DCOSGetVersion(w.client)
		if err != nil {
			return fmt.Errorf("Unable to get the DC/OS version: %s", err.Error())
		}

		rx := regexp.MustCompile(`([0-9]+\.[0-9]+)`)
		verParts := rx.FindStringSubmatch(vers.Version)
		if verParts == nil {
			return fmt.Errorf("Unable to parse the DC/OS version in: %s", vers.Version)
		}

		w.dcosVersion = verParts[1]
		log.Printf("detected=%s", w.dcosVersion)
	} else {
		log.Printf("cached version=%s", w.dcosVersion)
	}

	// If the cli does not exist in the sandbox, download now
	if w.cliBinary == "" {
		w.cliBinary = fmt.Sprintf("%s/bin/dcos-%s", w.sandbox, w.dcosVersion)
		if _, err := os.Stat(w.cliBinary); os.IsNotExist(err) {
			arch := runtime.GOARCH
			if arch == "amd64" {
				arch = "x86-64"
			}

			url := fmt.Sprintf("https://downloads.dcos.io/binaries/cli/%s/%s/dcos-%s/dcos", runtime.GOOS, arch, w.dcosVersion)
			log.Printf("dcos cli binary not found in sandbox, downloading from %s", url)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("Unable to download %s: %s", url, err.Error())
			}
			defer resp.Body.Close()

			// Create the file
			out, err := os.Create(w.cliBinary)
			if err != nil {
				return fmt.Errorf("Unable to create %s: %s", w.cliBinary, err.Error())
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return fmt.Errorf("Unable to download cli: %s", err.Error())
			}

			// Make sure it's executable
			err = os.Chmod(w.cliBinary, 0744)
			if err != nil {
				return fmt.Errorf("Unable to make downloaded cli executable")
			}
		}
		log.Printf("dcos cli binary found in %s", w.cliBinary)
	} else {
		log.Printf("cached binary=%s", w.cliBinary)
	}

	dcosConfig := w.client.CurrentDCOSConfig()

	// If the cluster is not configured, setup the cluster now
	clusterDir := fmt.Sprintf("%s/dcos/clusters/%s", w.sandbox, dcosConfig.ID())
	if _, err := os.Stat(clusterDir); os.IsNotExist(err) {
		log.Printf("dcos cluster was not configured in cli, configuring now")
		cmd := w.cliExec("cluster", "setup", "--insecure", "--no-check", dcosConfig.URL())
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("DCOS_CLUSTER_SETUP_ACS_TOKEN=%s", dcosConfig.ACSToken()),
		)
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Unable to setup cluster: %s", err.Error())
		}
		log.Printf("cluster configured")
	}

	// Always refresh the CLI token
	log.Printf("updating acs token")
	err := w.SetConfig("core.dcos_acs_token", dcosConfig.ACSToken())
	if err != nil {
		return fmt.Errorf("Unable to refresh cli token: %s", err.Error())
	}

	return nil
}

/**
 * Return an instance that is capable of executing package commands of the given package name
 */
func (w *CliWrapper) ForPackage(packageName string, id string) (*CliWrapperPackageWithConfig, error) {
	var pkg *CliWrapperPackage
	if inst, ok := w.instances[packageName]; ok {
		pkg = inst
	} else {
		pkg = &CliWrapperPackage{
			parent:         w,
			PackageName:    packageName,
			PackageCommand: "",
		}

		// Make sure the root package is prepared
		log.Printf("preparing sandbox")
		err := w.Prepare()
		if err != nil {
			return nil, err
		}

		// Prepare the cli-specific package
		err = pkg.Prepare()
		log.Printf("preparing package cli")
		if err != nil {
			return nil, fmt.Errorf("Unable to prepare cli for package %s: %s", packageName, err.Error())
		}

		// Cache this on the root structure in order to avoid costly
		// re-initializations by other resources
		w.instances[packageName] = pkg
	}

	return &CliWrapperPackageWithConfig{
		Package:    pkg,
		ID:         id,
		Config:     make(map[string]interface{}),
		LastOutput: "",
	}, nil
}

/**
 * Ensure that the package cli exists
 */
func (w *CliWrapperPackage) Prepare() error {
	log.Printf("Preparing package: %s", w.PackageName)

	if w.PackageCommand == "" {
		log.Printf("package command was not populated, detecting now")
		err := w.findPackageCliCommand()
		if err != nil {
			if _, ok := err.(*CliWrapperCommandNotFound); ok {
				// Ignore
			} else {
				if xerr, ok := err.(*exec.ExitError); ok {
					return fmt.Errorf("Unable to get package command: %s", string(xerr.Stderr))
				} else {
					return fmt.Errorf("Unable to get package command: %s", err.Error())
				}
			}
		}

		if w.PackageCommand == "" {
			log.Printf("cli package was not available, installing now")
			cmd := w.parent.cliExec("package", "install", "--yes", "--cli", w.PackageName)
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("Unable to  package: %s", err.Error())
			}
			err = cmd.Wait()
			if err != nil {
				return fmt.Errorf("Unable to  package: %s", err.Error())
			}

			log.Printf("checking if package is now installed")
			err = w.findPackageCliCommand()
			if err != nil {
				return fmt.Errorf("Unable to detect the installed package command: %s", err.Error())
			}
		}

		if w.PackageCommand == "" {
			return fmt.Errorf("Unable to detect the package command")
		}
	}

	return nil
}

/**
 * Ensure that the package cli exists
 */
func (w *CliWrapperPackage) findPackageCliCommand() error {
	ret, err := w.parent.cliExec("package", "list", "--json", w.PackageName).Output()
	if err != nil {
		return err
	}

	var resp []PackageListResonse
	err = json.Unmarshal(ret, &resp)
	if err != nil {
		return fmt.Errorf("Unable to parse response: %s", err.Error())
	}

	if len(resp) == 0 {
		return &CliWrapperCommandNotFound{"Package was not found"}
	}

	log.Printf("Parsd list json='%s', to: %d", string(ret), resp)
	if resp[0].Command != nil {
		w.PackageCommand = resp[0].Command.Name
	}

	return nil
}

/**
 * Invoke the given command-line and pipe the configuration to the launched
 * process, according to the command-line given:
 *
 * - If the expression contains "< %CONFIG%", the config JSON will be piped
 *   to the STDIN of the process.
 * - If the expression contains "> %CONFIG%", the config JSON will be piped
 *   out from the STDOUT of the process.
 * - If the expression contains "%CONFIG%", a temporary file will be created
 *   and will replace the macro placeholder.
 */
func (w *CliWrapperPackageWithConfig) Exec(argline string, shell bool) error {
	var stdout io.ReadCloser
	var stderr io.ReadCloser
	var stdin io.WriteCloser

	var tempFile *os.File
	var tempFileError error

	var cmd *exec.Cmd
	var useFile bool = false
	var useStdin bool = false
	var useStdout bool = false

	var replaceCb = func(n string) string {
		if n[0] == '<' {
			useStdin = true
			return ""
		}
		if n[0] == '>' {
			useStdout = true
			return ""
		}

		useFile = true
		tempFile, tempFileError = ioutil.TempFile("", "config")
		if tempFileError != nil {
			return ""
		}

		return tempFile.Name()
	}

	// replace %CONIG%, %ID% and break into args
	argline = strings.Trim(
		rxConfigArg.ReplaceAllStringFunc(
			rxIDArg.ReplaceAllString(
				argline, w.ID,
			),
			replaceCb),
		"\t\r\n ",
	)
	if tempFileError != nil {
		return fmt.Errorf("Could not create temporary file: %s", tempFileError.Error())
	}

	// Pass it down to shell
	if shell {
		cmd = w.Package.parent.shellExec(argline)
	} else {
		args := strings.Split(
			argline,
			" ",
		)
		args = append([]string{w.Package.PackageCommand}, args...)
		cmd = w.Package.parent.cliExec(args...)
	}

	// Get the configuration contents
	configBytes, err := json.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("Unable to generate JSON config: %s", err.Error())
	}

	// Either open STDIN or STDOUT depending on if we are going to read
	// or write towards the process stream
	if useStdin {
		stdin, err = cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to open stdin pipe: %s", err.Error())
		}
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to open stdout pipe: %s", err.Error())
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to open stderr pipe: %s", err.Error())
	}

	// If we are using a config file, dump the contents now
	if useFile {
		tempFile.Write(configBytes)
		tempFile.Close()
		defer os.Remove(tempFile.Name())
	}

	// Launch process
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Unable to start process: %s", err.Error())
	}

	// If we should push STDIN, push now
	if useStdin {
		stdin.Write(configBytes)
		stdin.Close()
	}

	// If we should collect from STDOUT, read now
	if useStdout {
		configBytes, _ = ioutil.ReadAll(stdout)
		stdout.Close()
	} else {
		outBytes, _ := ioutil.ReadAll(stdout)
		stdout.Close()
		w.LastOutput = strings.Trim(string(outBytes), " \t\r\n")
	}

	// Read stderr
	errBytes, _ := ioutil.ReadAll(stderr)
	defer stderr.Close()
	if w.LastOutput != "" {
		w.LastOutput += "\n"
	}
	w.LastOutput += strings.Trim(string(errBytes), " \t\r\n")

	// Wait for completion
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// If we are using a config file, read the config contents now
	if useFile {
		configBytes, err = ioutil.ReadFile(tempFile.Name())
		if err != nil {
			return fmt.Errorf("Unable to read config file: %s", err.Error())
		}
	}

	// If the config has been potentially updated, re-load
	if useFile || useStdout {
		log.Printf("Collected config=%s", string(configBytes))

		// Check if we can read it as dict
		err = json.Unmarshal(configBytes, &w.Config)
		if err != nil {
			log.Printf("Parsing as map[string]interface{} failed, trying list")
			var dictList []map[string]interface{}

			// Check if we can read it as list of dict
			err = json.Unmarshal(configBytes, &dictList)
			if err != nil {
				log.Printf("Parsing as []map[string]interface{} failed: %s", err.Error())
				return &CliWrapperConfigParseError{fmt.Sprintf("Unable to reload the config: %s", err.Error())}
			} else {
				if len(dictList) == 0 {
					return &CliWrapperConfigParseError{fmt.Sprintf("Got empty config: %s", err.Error())}
				} else {
					w.Config = dictList[0]
				}
			}
		}
	}

	return nil
}
