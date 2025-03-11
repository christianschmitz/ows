package resources

import (
	"errors"
	"os/exec"
	"ows/ledger"
)

type TaskConfig struct {
	Runtime string
	Handler ledger.AssetId
}

type TasksManager struct {
	Tasks map[string]TaskConfig
}

func NewTasksManager() *TasksManager {
	return &TasksManager{
		map[string]TaskConfig{},
	}
}

func (m *TasksManager) Add(id ledger.ResourceId, handler ledger.AssetId) error {
	sId := ledger.StringifyResourceId(id)

	if _, ok := m.Tasks[sId]; ok {
		return errors.New("task added before")
	}

	m.Tasks[sId] = TaskConfig{"nodejs", handler}

	return nil
}

func (m *TasksManager) Run(id ledger.ResourceId) (string, error) {
	sId := ledger.StringifyResourceId(id)
	taskConf, ok := m.Tasks[sId]

	if !ok {
		return "", errors.New("task not found")
	}

	entryPointPath := ledger.HomeDir + "/assets/" + ledger.StringifyAssetId(taskConf.Handler)

	if (taskConf.Runtime == "nodejs") {
		// run the task using plain nodejs initially
		// TODO: Docker
		return runNodeScript(entryPointPath)
	} else {
		return "", errors.New("unsupported runtime " + taskConf.Runtime)
	}
}

func runNodeScript(scriptPath string) (string, error) {
	cmd := exec.Command("node", scriptPath)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}