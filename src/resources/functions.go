package resources

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"

	"ows/ledger"
)

const DOCKER_IMAGE_NAME = "ows_nodejs_image"
const DOCKER_CONTAINER_NAME = "ows_nodejs_container"
const NODEJS_RUNNER_NAME = "runner.js"
const NODEJS_HANDLER_NAME = "handler.js"
const NODEJS_INPUT_NAME = "input.json"
const NODEJS_OUTPUT_NAME = "output.json"
const IPC_SOCKET_NAME = "socket.sock"

type Function struct {
	Config ledger.FunctionConfig
}

type FunctionManager struct {
	dockerInitialized bool
	Assets            *AssetManager
	Functions         map[ledger.FunctionID]*Function
}

func newFunctionManager(assets *AssetManager) *FunctionManager {
	return &FunctionManager{
		dockerInitialized: false,
		Assets:            assets,
		Functions:         map[ledger.FunctionID]*Function{},
	}
}

func (m *FunctionManager) Sync(functions map[ledger.FunctionID]ledger.FunctionConfig) error {
	for id, conf := range functions {
		if _, ok := m.Functions[id]; ok {
			if err := m.update(id, conf); err != nil {
				return fmt.Errorf("failed to update function %s (%v)", id, err)
			}
		} else {
			if err := m.add(id, conf); err != nil {
				return fmt.Errorf("failed to add function %s (%v)", id, err)
			}
		}
	}

	for id, _ := range m.Functions {
		if _, ok := functions[id]; !ok {
			if err := m.remove(id); err != nil {
				return fmt.Errorf("failed to remove function %s (%v)", id, err)
			}
		}
	}

	return nil
}

func (m *FunctionManager) add(id ledger.FunctionID, config ledger.FunctionConfig) error {
	log.Printf("adding function %s with handler %s...\n", id, config.HandlerID)

	if _, ok := m.Functions[id]; ok {
		return errors.New("function added before")
	}

	if !m.dockerInitialized {
		if err := m.initializeDocker(); err != nil {
			log.Println("failed to initialize docker", err)
			return err
		}

		m.dockerInitialized = true
	}

	if err := m.Assets.AssertExists(config.HandlerID); err != nil {
		return err
	}

	m.Functions[id] = &Function{
		Config: config,
	}

	return nil
}

func (m *FunctionManager) remove(id ledger.FunctionID) error {
	if _, ok := m.Functions[id]; !ok {
		return errors.New("function not found")
	}

	// TODO: what if task is still being referenced elsewhere?

	delete(m.Functions, id)

	return nil
}

func (m *FunctionManager) update(id ledger.FunctionID, config ledger.FunctionConfig) error {
	fn, ok := m.Functions[id]
	if !ok {
		return fmt.Errorf("function %s not found", id)
	}

	fn.Config = config

	return nil
}

func (m *FunctionManager) Run(id ledger.FunctionID, arg any) (any, error) {
	fn, ok := m.Functions[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}

	conf := fn.Config

	if conf.Runtime != "nodejs" {
		return nil, fmt.Errorf("unsupported runtime %s", conf.Runtime)
	}

	return m.runNodeScriptInDocker(string(conf.HandlerID), arg)
}

func makeTmpDir() (string, error) {
	tmpDir := "/tmp/" + uuid.NewString()

	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return "", err
	}

	return tmpDir, nil
}

func runNodeScriptDirectly(scriptPath string) (string, error) {
	cmd := exec.Command("node", scriptPath)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (m *FunctionManager) runNodeScriptInDocker(handler string, arg any) (any, error) {
	start := time.Now()

	tmpDir, err := makeTmpDir()
	if err != nil {
		return nil, err
	}

	dirFields := strings.Split(tmpDir, "/")
	uuid := dirFields[len(dirFields)-1]

	fmt.Println("running task " + tmpDir)

	defer os.RemoveAll(tmpDir)

	// write the necessary files
	if err := m.copyAsset(handler, tmpDir+"/"+NODEJS_HANDLER_NAME); err != nil {
		return nil, err
	}

	if err := writeJson(arg, tmpDir+"/"+NODEJS_INPUT_NAME); err != nil {
		fmt.Println("failed to write input")
		return nil, err
	}

	startInner := time.Now()

	// send a request to the socker
	conn, err := net.Dial("unix", "/tmp/"+IPC_SOCKET_NAME)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	_, err = conn.Write([]byte(uuid + "\n"))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	durationInner := time.Since(startInner)

	var output map[string]any
	err = json.Unmarshal([]byte(response), &output)
	if err != nil {
		return nil, err
	}

	log.Printf("task %s took %s (inner), %s (outer)\n", uuid, durationInner, time.Since(start))

	if s, ok := output["success"]; ok {
		if b, ok := s.(bool); ok {
			if !b {
				if e, ok := output["error"]; ok {
					if em, ok := e.(string); ok {
						return nil, errors.New(em)
					} else {
						return nil, errors.New("unexpected output format")
					}
				} else {
					return nil, errors.New("unexpected output format")
				}
			} else {
				if r, ok := output["result"]; ok {
					return r, nil
				} else {
					return nil, errors.New("unexpected output format")
				}
			}
		} else {
			return nil, errors.New("unexpected output format")
		}
	} else {
		return nil, errors.New("unexpected output format")
	}
}

func (m *FunctionManager) copyAsset(assetId string, dst string) error {
	assetsDir := m.Assets.AssetsDir
	src := assetsDir + "/" + assetId

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dst, input, 0644); err != nil {
		return err
	}

	return nil
}

func writeJson(input any, dst string) error {
	inputData, err := json.Marshal(input)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dst, inputData, 0644); err != nil {
		return err
	}

	return nil
}

func readJson(src string) (any, error) {
	outputData, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}

	var output any

	if err := json.Unmarshal(outputData, &output); err != nil {
		return nil, err
	}

	return output, nil
}

// TODO: take runtime into account
func (m *FunctionManager) dockerContainerName() string {
	return DOCKER_CONTAINER_NAME + string(m.Assets.Nodes.CurrentNodeID())
}

func (m *FunctionManager) initializeDocker() error {
	// check if container is already running first
	cmd := exec.Command("docker", "ps", "-q", "-f", "name="+m.dockerContainerName())
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Println(out)
		return err
	}

	if out.Len() != 0 {
		if err := exec.Command("docker", "stop", m.dockerContainerName()).Run(); err != nil {
			return err
		}
	}

	log.Println("creating and starting nodejs runner...")

	// make sure the previous container (which might be stopped), is completely removed
	exec.Command("docker", "container", "rm", "-f", m.dockerContainerName()).Run()

	// make sure Docker image exists
	tmpDir, err := makeTmpDir()
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	if err := writeNodejsDockerfile(tmpDir + "/Dockerfile"); err != nil {
		return err
	}

	if err := writeNodejsRunner2(tmpDir + "/" + NODEJS_RUNNER_NAME); err != nil {
		return err
	}

	log.Println("created docker files")

	cmd = exec.Command("docker", "build", "-t", DOCKER_IMAGE_NAME, tmpDir)

	if _, err := cmd.Output(); err != nil {
		return err
	}

	cmd = exec.Command("docker", "run", "-d", "--name", m.dockerContainerName(), "-v", "/tmp:/data", DOCKER_IMAGE_NAME)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if output, err := cmd.Output(); err != nil {
		fmt.Println("Stderr: ", stderr.String())
		fmt.Println("Output: ", output)
		return err
	}

	return nil
}

func writeNodejsDockerfile(dst string) error {
	dockerfileLines := []string{
		"FROM node:18-alpine",
		"WORKDIR /data",
		"COPY " + NODEJS_RUNNER_NAME + " /app/" + NODEJS_RUNNER_NAME,
		//"CMD exec /bin/sh -c \"trap : TERM INT; sleep infinity & wait\"",
		"CMD node /app/runner.js",
	}

	dockerfile := strings.Join(dockerfileLines, "\n")

	return ioutil.WriteFile(dst, []byte(dockerfile), 0644)
}

func writeNodejsRunner(dst string) error {
	// common js
	// this is not the approach with the least overhead, but seems to work fine
	runnerLines := []string{
		"const {promises: fs} = require('fs');",
		"const path = require('path');",
		"const taskId = process.argv[process.argv.length - 1];",
		"const inputFilePath = '/data/' + taskId + '/" + NODEJS_INPUT_NAME + "';",
		"const outputFilePath = '/data/' + taskId + '/" + NODEJS_OUTPUT_NAME + "';",
		"const handlerFilePath = '/data/' + taskId + '/" + NODEJS_HANDLER_NAME + "';",
		"async function main() {",
		"    try {",
		"        const inputData = await fs.readFile(inputFilePath, 'utf-8');",
		"        const input = JSON.parse(inputData);",
		"        const handler = require(handlerFilePath);",
		"        let output = handler(input);",
		"		 if (output instanceof Promise) {",
		"            output = await output;",
		"        }",
		"        await fs.writeFile(outputFilePath, JSON.stringify(output, null, 2));",
		"        console.log('Task completed successfully.');",
		"    } catch (error) {",
		"        console.error('Error:', error);",
		"        process.exit(1);",
		"    }",
		"}",
		"main();",
	}

	runner := strings.Join(runnerLines, "\n")

	return ioutil.WriteFile(dst, []byte(runner), 0644)
}

// the runner starts a socket for IPC, which it owns
// the go process sends the task ID through the socket
// the runner returns the JSON output
// the docker process runs with super-user rights, so the socket must be created for all to be able to access
func writeNodejsRunner2(dst string) error {
	runnerLines := []string{
		"const {promises: fs} = require('fs');",
		"const path = require('path');",
		"const net = require('net');",
		"const { exec } = require('child_process');",
		"const socketPath = '/data/" + IPC_SOCKET_NAME + "';",
		"async function run() {",
		"try {await fs.unlink(socketPath);}catch(err){}",
		"const server = net.createServer(async (socket) => {",
		"    socket.on('data', async (data) => {",
		"        const taskId = data.toString().trim();",
		"        console.log('Processing task: ' + taskId);",
		"        try {",
		"            const inputFilePath = path.join('/data', taskId, '" + NODEJS_INPUT_NAME + "');",
		"            const inputData = await fs.readFile(inputFilePath, 'utf-8');",
		"            const input = JSON.parse(inputData);",
		"            const handlerFilePath = path.join('/data', taskId, '" + NODEJS_HANDLER_NAME + "');",
		"            const handler = require(handlerFilePath);",
		"            let output = handler(input);",
		"		     if (output instanceof Promise) {",
		"                output = await output;",
		"            }",
		"            socket.write(JSON.stringify({success: true, result: output}) + '\\n');",
		"        } catch (err) {",
		"            socket.write(JSON.stringify({success: false, error: err.message}) + '\\n')",
		"        }",
		"    })",
		"})",
		"server.listen(socketPath, () => {",
		"    exec('chmod 777 /data/" + IPC_SOCKET_NAME + "', (err) => {",
		"        if (err) {",
		"            console.error('Failed to change socket permissions:', err);",
		"        } else {",
		"            console.log('Socket permissions changed to 777 (read/write for everyone)');",
		"        }",
		"    });",
		"})",
		"}",
		"run();",
	}

	runner := strings.Join(runnerLines, "\n")

	return ioutil.WriteFile(dst, []byte(runner), 0644)
}
