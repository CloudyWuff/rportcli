package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/config"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestScriptExecutionByClientIDsSuccess(t *testing.T) {
	const scriptToExecute = "cwd"
	scriptToExecuteBase64 := base64.StdEncoding.EncodeToString([]byte(scriptToExecute))

	testCases := []struct {
		scriptPathToGive   string
		name               string
		interpreterToGive  string
		commandToExpect    *models.WsScriptCommand
		errorToExpect      string
		shouldCreateScript bool
	}{
		{
			scriptPathToGive:   "some_powershell_script.ps1",
			shouldCreateScript: true,
			name:               "exec_powershell_script",
			commandToExpect: &models.WsScriptCommand{
				ClientIDs:           []string{"2222"},
				IsSudo:              false,
				ExecuteConcurrently: false,
				AbortOnError:        false,
				TimeoutSec:          DefaultCmdTimeoutSeconds,
				Command:             "",
				Script:              scriptToExecuteBase64,
				Cwd:                 "/home",
				Interpreter:         "powershell",
			},
		},
		{
			scriptPathToGive:   "some_cmd_script.bat",
			name:               "exec_cmd_script",
			shouldCreateScript: true,
			commandToExpect: &models.WsScriptCommand{
				ClientIDs:   []string{"2223"},
				TimeoutSec:  DefaultCmdTimeoutSeconds,
				Script:      scriptToExecuteBase64,
				Cwd:         "/home2",
				Interpreter: "cmd",
			},
		},
		{
			scriptPathToGive:   "some_cmd_script.bat",
			name:               "exec_script_from_param",
			shouldCreateScript: true,
			commandToExpect: &models.WsScriptCommand{
				ClientIDs:   []string{"2224"},
				TimeoutSec:  DefaultCmdTimeoutSeconds,
				Script:      scriptToExecuteBase64,
				Cwd:         "/home3",
				Interpreter: "powershell",
			},
			interpreterToGive: "powershell",
		},
		{
			scriptPathToGive:   "some_cmd_script",
			name:               "empty_script_file_ext",
			shouldCreateScript: true,
			commandToExpect: &models.WsScriptCommand{
				ClientIDs:   []string{"2225"},
				TimeoutSec:  DefaultCmdTimeoutSeconds,
				Script:      scriptToExecuteBase64,
				Cwd:         "/home4",
				Interpreter: "",
			},
			interpreterToGive: "",
		},
		{
			scriptPathToGive:   "some_cmd_script.txt",
			name:               "unknown_script_file_ext",
			shouldCreateScript: true,
			commandToExpect: &models.WsScriptCommand{
				ClientIDs:   []string{"2226"},
				TimeoutSec:  DefaultCmdTimeoutSeconds,
				Script:      scriptToExecuteBase64,
				Cwd:         "/home6",
				Interpreter: "",
			},
			interpreterToGive: "",
		},
		{
			scriptPathToGive:   "some_unknown_script.sh",
			name:               "non_existing_script_path",
			shouldCreateScript: false,
			errorToExpect:      "script file doesn't exist: some_unknown_script.sh",
		},
		{
			scriptPathToGive:   "",
			name:               "empty_existing_script_path",
			shouldCreateScript: false,
			errorToExpect:      "required option script is empty",
		},
	}

	tmpDir := os.TempDir()
	writtenFiles := []string{}
	defer func() {
		for _, writtenFile := range writtenFiles {
			err := os.Remove(writtenFile)
			if err != nil {
				log.Errorln(err)
			}
		}
	}()

	for _, testCase := range testCases {
		tc := testCase
		t.Run(testCase.name, func(t *testing.T) {
			scriptPath := ""
			if tc.shouldCreateScript {
				scriptPath = filepath.Join(tmpDir, tc.scriptPathToGive)
				err := ioutil.WriteFile(scriptPath, []byte(scriptToExecute), 0600)
				require.NoError(t, err)
				writtenFiles = append(writtenFiles, scriptPath)
			} else {
				scriptPath = tc.scriptPathToGive
			}

			params := map[string]string{
				Interpreter: tc.interpreterToGive,
				Script:      scriptPath,
			}

			if tc.commandToExpect != nil {
				params[ClientIDs] = strings.Join(tc.commandToExpect.ClientIDs, ",")
				params[Cwd] = tc.commandToExpect.Cwd
			}

			paramsContainer := config.FromValues(params)

			jobToGive := buildJob()
			sc, rw, jr, err := buildScriptController(jobToGive)
			require.NoError(t, err)

			err = sc.Start(context.Background(), paramsContainer)
			if tc.errorToExpect != "" {
				require.Contains(t, err.Error(), tc.errorToExpect)
				return
			}

			require.NoError(t, err)

			assert.Len(t, rw.writtenItems, 1)
			actualWrittenScriptInput := &models.WsScriptCommand{}
			err = json.Unmarshal([]byte(rw.writtenItems[0]), actualWrittenScriptInput)
			require.NoError(t, err)

			assert.Equal(t, tc.commandToExpect.Command, actualWrittenScriptInput.Command)
			assert.Equal(t, tc.commandToExpect.Script, actualWrittenScriptInput.Script)
			assert.Equal(t, tc.commandToExpect.ClientIDs, actualWrittenScriptInput.ClientIDs)
			assert.Equal(t, tc.commandToExpect.Cwd, actualWrittenScriptInput.Cwd)
			assert.Equal(t, tc.commandToExpect.Interpreter, actualWrittenScriptInput.Interpreter)
			assert.Equal(t, tc.commandToExpect.GroupIDs, actualWrittenScriptInput.GroupIDs)
			assert.Equal(t, tc.commandToExpect.IsSudo, actualWrittenScriptInput.IsSudo)
			assert.Equal(t, tc.commandToExpect.TimeoutSec, actualWrittenScriptInput.TimeoutSec)
			assert.Equal(t, tc.commandToExpect.ExecuteConcurrently, actualWrittenScriptInput.ExecuteConcurrently)
			assert.Equal(t, tc.commandToExpect.AbortOnError, actualWrittenScriptInput.AbortOnError)

			assert.True(t, rw.isClosed)
			require.NotNil(t, jr.jobToRender)
			assert.Equal(t, jobToGive.Command, jr.jobToRender.Command)
			assert.Equal(t, jobToGive.IsScript, jr.jobToRender.IsScript)
			assert.Equal(t, jobToGive.ClientID, jr.jobToRender.ClientID)
			assert.Equal(t, jobToGive.Cwd, jr.jobToRender.Cwd)
			assert.Equal(t, jobToGive.Interpreter, jr.jobToRender.Interpreter)
			assert.Equal(t, jobToGive.IsSudo, jr.jobToRender.IsSudo)
			assert.Equal(t, jobToGive.TimeoutSec, jr.jobToRender.TimeoutSec)
			assert.Equal(t, jobToGive.Error, jr.jobToRender.Error)
			assert.Equal(t, jobToGive.Status, jr.jobToRender.Status)
			assert.Equal(t, jobToGive.ClientName, jr.jobToRender.ClientName)
			assert.Equal(t, jobToGive.Jid, jr.jobToRender.Jid)
			assert.Equal(t, jobToGive.MultiJobID, jr.jobToRender.MultiJobID)
			assert.Equal(t, jobToGive.CreatedBy, jr.jobToRender.CreatedBy)
			assert.Equal(t, jobToGive.Result, jr.jobToRender.Result)
		})
	}
}

func buildJob() *models.Job {
	return &models.Job{
		Jid:        "934",
		Status:     "in_progress",
		FinishedAt: time.Now(),
		ClientID:   "2222",
		Command:    "pwd",
		StartedAt:  time.Now(),
		CreatedBy:  "admin",
		TimeoutSec: 1,
		Result: models.JobResult{
			Stdout: "some out",
			Stderr: "some err",
		},
	}
}

func buildScriptController(j *models.Job) (*ScriptsController, *ReadWriterMock, *JobRendererMock, error) {
	jobRespBytes, err := json.Marshal(j)
	if err != nil {
		return nil, nil, nil, err
	}

	rw := &ReadWriterMock{
		itemsToRead: []ReadChunk{
			{
				Output: jobRespBytes,
			},
			{
				Err: io.EOF,
			},
		},
		writtenItems: []string{},
		isClosed:     false,
	}

	jr := &JobRendererMock{}

	return &ScriptsController{
		ExecutionHelper: &ExecutionHelper{
			ReadWriter:  rw,
			JobRenderer: jr,
		},
	}, rw, jr, nil
}
