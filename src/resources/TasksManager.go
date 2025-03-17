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
)

const DOCKER_IMAGE_NAME = "ows_nodejs_image"
const DOCKER_CONTAINER_NAME = "ows_nodejs_container"
const NODEJS_RUNNER_NAME = "runner.js"
const NODEJS_HANDLER_NAME = "handler.js"
const NODEJS_INPUT_NAME = "input.json"
const NODEJS_OUTPUT_NAME = "output.json"
const IPC_SOCKET_NAME = "socket.sock"

type TaskConfig struct {
	Runtime string
	Handler string
}

type TasksManager struct {
	dockerInitialized bool
	Tasks             map[string]TaskConfig
}

func NewTasksManager() *TasksManager {
	return &TasksManager{
		false,
		map[string]TaskConfig{},
	}
}

func (m *TasksManager) Add(id string, handler string) error {
	if _, ok := m.Tasks[id]; ok {
		return errors.New("task added before")
	}

	if !m.dockerInitialized {
		if err := initializeDocker(); err != nil {
			fmt.Println("failed to initialize docker", err)
			return err
		}

		m.dockerInitialized = true
	}

	m.Tasks[id] = TaskConfig{"nodejs", handler}

	fmt.Println("added task " + id)
	return nil
}

func (m *TasksManager) Remove(id string) error {
	if _, ok := m.Tasks[id]; !ok {
		return errors.New("task not found")
	}

	// TODO: what if task is still being referenced elsewhere?

	delete(m.Tasks, id)

	return nil
}

func makeTmpDir() (string, error) {
	tmpDir := HomeDir + "/tmp/" + uuid.NewString()

	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return "", err
	}

	return tmpDir, nil
}

func (m *TasksManager) Run(id string, arg any) (any, error) {
	taskConf, ok := m.Tasks[id]
	if !ok {
		return nil, errors.New("task " + id + " not found")
	}

	if taskConf.Runtime != "nodejs" {
		return nil, errors.New("unsupported runtime " + taskConf.Runtime)
	}

	return runNodeScriptInDocker(taskConf.Handler, arg)
}

func runNodeScriptDirectly(scriptPath string) (string, error) {
	cmd := exec.Command("node", scriptPath)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func runNodeScriptInDocker(handler string, arg any) (any, error) {
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
	if err := copyAsset(handler, tmpDir+"/"+NODEJS_HANDLER_NAME); err != nil {
		return nil, err
	}

	if err := writeJson(arg, tmpDir+"/"+NODEJS_INPUT_NAME); err != nil {
		fmt.Println("failed to write input")
		return nil, err
	}

	startInner := time.Now()

	// run the docker container
	// TODO: send command to IPC socket instead
	conn, err := net.Dial("unix", HomeDir+"/tmp/"+IPC_SOCKET_NAME)
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

	//cmd := exec.Command("docker", "exec", DOCKER_CONTAINER_NAME, "node", "/app/" + NODEJS_RUNNER_NAME, uuid)
	//
	//var stderr bytes.Buffer
	//cmd.Stderr = &stderr
	//
	//// TODO: write stdout and stderr to logs
	//if output, err := cmd.Output(); err != nil {
	//	fmt.Println(err)
	//	fmt.Println("Stderr: ", stderr.String())
	//	fmt.Println("Output: ", output)
	//	return nil, err
	//}
	//
	//
	//
	//
	//fmt.Println("done running, reading output")
	//
	//output, err := readJson(tmpDir + "/" + NODEJS_OUTPUT_NAME)
	//if err != nil {
	//	fmt.Println("failed to read output file")
	//	return nil, err
	//}
	//
	//log.Printf("task %s took %s (inner), %s (outer)\n", uuid, durationInner, time.Since(start))
	//
	//return output, nil
}

func copyAsset(assetId string, dst string) error {
	assetsDir := GetAssetsDir()
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

func initializeDocker() error {
	// check if container is already running first
	cmd := exec.Command("docker", "ps", "-q", "-f", "name="+DOCKER_CONTAINER_NAME)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Println(out)
		return err
	}

	if out.Len() != 0 {
		if err := exec.Command("docker", "stop", DOCKER_CONTAINER_NAME).Run(); err != nil {
			return err
		}
	}

	fmt.Println("Creating and starting nodejs runner...")

	// make sure the previous container (which might be stopped), is completely removed
	exec.Command("docker", "container", "rm", "-f", DOCKER_CONTAINER_NAME).Run()

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

	fmt.Println("Created docker files")

	cmd = exec.Command("docker", "build", "-t", DOCKER_IMAGE_NAME, tmpDir)

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("HERE:", output)
		return err
	}

	cmd = exec.Command("docker", "run", "-d", "--name", DOCKER_CONTAINER_NAME, "-v", HomeDir+"/tmp:/data", DOCKER_IMAGE_NAME)

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
		"            socket.write(JSON.stringify({success: false, error: error.message}) + '\\n')",
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
