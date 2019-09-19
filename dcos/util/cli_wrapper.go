package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/dcos/client-go/dcos"
)

var rxConfigArg = regexp.MustCompile(`([><] *)?%CONFIG%`)
var rxIDArg = regexp.MustCompile(`%ID%`)

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

func CreateCliWrapper(sandbox string, client *dcos.APIClient) (*CliWrapper, error) {
	return &CliWrapper{
		sandbox:     sandbox,
		client:      client,
		cliBinary:   "",
		dcosVersion: "",
		instances:   make(map[string]*CliWrapperPackage),
	}, nil
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
	Package *CliWrapperPackage
	ID      string
	Config  map[string]interface{}
}

/**
 * Get a configuration property
 */
func (w *CliWrapper) GetConfig(name string) (string, error) {
	cmd := exec.Command(w.cliBinary, "config", "show", name)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", w.sandbox),
	)

	ret, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

/**
 * Update a configuration property
 */
func (w *CliWrapper) SetConfig(name string, value string) error {
	cmd := exec.Command(w.cliBinary, "config", "set", name, value)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", w.sandbox),
	)

	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

/**
 * Ensure that the cli for the correct DC/OS version exists in the sandbox
 */
func (w *CliWrapper) Prepare() error {

	// Ensure sandbox structure
	os.MkdirAll(fmt.Sprintf("%s/bin", w.sandbox), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/home", w.sandbox), os.ModePerm)

	// If the DC/OS version is not defined, pull the version now
	if w.dcosVersion == "" {
		verExpr, err := DCOSGetVersion(w.client)
		if err != nil {
			return fmt.Errorf("Unable to get the DC/OS version: %s", err.Error())
		}

		rx := regexp.MustCompile(`dcos-([0-9]+\.[0-9]+)`)
		verParts := rx.FindStringSubmatch(verExpr)
		if verParts == nil {
			return fmt.Errorf("Unable to parse the DC/OS version")
		}

		w.dcosVersion = verParts[1]
	}

	// If the cli does not exist in the sandbox, download now
	if w.cliBinary == "" {
		w.cliBinary = fmt.Sprintf("%s/bin/dcos-%s", w.sandbox, w.dcosVersion)
		if _, err := os.Stat(w.cliBinary); os.IsNotExist(err) {
			// Get the data
			url := fmt.Sprintf("https://downloads.dcos.io/binaries/cli/%s/%s/%s/dcos",
				runtime.GOOS, runtime.GOARCH, w.dcosVersion)
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
	}

	dcosConfig := w.client.CurrentDCOSConfig()

	// If the cluster is not configured, setup the cluster now
	clusterDir := fmt.Sprintf("%s/home/.dcos/clusters/%s", w.sandbox, dcosConfig.ID())
	if _, err := os.Stat(clusterDir); os.IsNotExist(err) {
		cmd := exec.Command(w.cliBinary, "cluster", "setup", "--insecure", "--no-check", dcosConfig.URL())
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("HOME=%s", w.sandbox),
			fmt.Sprintf("DCOS_CLUSTER_SETUP_ACS_TOKEN=%s", dcosConfig.ACSToken()),
		)
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Unable to setup cluster: %s", err.Error())
		}
	}

	// Always refresh the CLI token
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
		err := w.Prepare()
		if err != nil {
			return nil, err
		}

		// Prepare the cli-specific package
		err = pkg.Prepare()
		if err != nil {
			return nil, fmt.Errorf("Unable to prepare cli for package %s: %s", packageName, err.Error())
		}

		// Cache this on the root structure in order to avoid costly
		// re-initializations by other resources
		w.instances[packageName] = pkg
	}

	return &CliWrapperPackageWithConfig{
		Package: pkg,
		ID:      id,
		Config:  make(map[string]interface{}),
	}, nil
}

/**
 * Ensure that the package cli exists
 */
func (w *CliWrapperPackage) Prepare() error {
	// TODO: Install package sub-command
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
func (w *CliWrapperPackageWithConfig) Exec(cmdline string) error {
	var stdout io.ReadCloser
	var stdin io.WriteCloser

	var tempFile *os.File
	var tempFileError error

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
	args := strings.Split(
		strings.Trim(
			rxConfigArg.ReplaceAllStringFunc(
				rxIDArg.ReplaceAllString(
					cmdline, w.ID,
				),
				replaceCb),
			"\t\r\n ",
		),
		" ",
	)

	if tempFileError != nil {
		return fmt.Errorf("Could not create temporary file: %s", tempFileError.Error())
	}

	// Get the configuration contents
	configBytes, err := json.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("Unable to generate JSON config: %s", err.Error())
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", w.Package.parent.sandbox),
	)

	// Either open STDIN or STDOUT depending on if we are going to read
	// or write towards the process stream
	if useStdin {
		stdin, err = cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to open stdin pipe: %s", err.Error())
		}
	}
	if useStdout {
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to open stdout pipe: %s", err.Error())
		}
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
	}

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
		err = json.Unmarshal(configBytes, &w.Config)
		if err != nil {
			return fmt.Errorf("Unable to reload the config: %s", err.Error())
		}
	}

	return nil
}
