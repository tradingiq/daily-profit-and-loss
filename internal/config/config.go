package config

import (
	"context"
	pnlapp "daily-profit-and-loss/internal/app"
	"daily-profit-and-loss/internal/logger"
	"daily-profit-and-loss/internal/ui"
	"encoding/json"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/tradingiq/bitunix-client/bitunix"
	"github.com/tradingiq/bitunix-client/model"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ApiKey            string        `json:"api_key"`
	SecretKey         string        `json:"secret_key"`
	ProfitAndLossFile string        `json:"profit_and_loss_file"`
	Mtx               sync.Mutex    `json:"-"`
	Changed           chan struct{} `json:"-"`
}

func LoadConfig() *Config {
	log := logger.GetInstance()

	config := &Config{
		Changed: make(chan struct{}),
	}

	configPath := pnlapp.GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("could not read config file (this is normal for first run):", err)
		return config
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Error("could not parse config file:", err)
		return &Config{}
	}

	return config
}

func SaveConfig(config *Config) error {
	log := logger.GetInstance()
	config.Mtx.Lock()
	defer config.Mtx.Unlock()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Error("error marshaling config:", err)
		return fmt.Errorf("error marshaling config: %w", err)
	}

	configPath := pnlapp.GetConfigPath()
	return os.WriteFile(configPath, data, 0600)
}

func RunConfigWindow(w *app.Window, config *Config, log *logger.Logger) error {
	th := material.NewTheme()

	var (
		apiKeyInput     widget.Editor
		secretKeyInput  widget.Editor
		folderPathInput widget.Editor
		selectFolderBtn widget.Clickable
		saveButton      widget.Clickable
		closeButton     widget.Clickable
	)

	apiKeyInput.SingleLine = true
	secretKeyInput.SingleLine = true
	folderPathInput.SingleLine = true

	apiKeyInput.SetText(config.ApiKey)
	secretKeyInput.SetText(config.SecretKey)
	folderPathInput.SetText(config.ProfitAndLossFile)

	status := ""

	configHandler := func(gtx layout.Context, theme interface{}, closeRequested chan bool) layout.Dimensions {
		th := theme.(*material.Theme)

		ui.CloseButtonHandler(&closeButton, gtx, closeRequested)

		if selectFolderBtn.Clicked(gtx) {

			go func() {
				selectedPath := ShowFolderPicker(log)
				if selectedPath != "" {

					folderPathInput.SetText(selectedPath)
					status = "Folder selected: " + selectedPath
				} else {
					status = "Folder selection canceled or failed"
				}
			}()
		}

		if saveButton.Clicked(gtx) {
			apiKey := apiKeyInput.Text()
			secretKey := secretKeyInput.Text()
			folderPath := folderPathInput.Text()

			apiClient, err := bitunix.NewApiClient(apiKey, secretKey)
			if err != nil {
				log.Error("failed to create API client: %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()

			if _, err := apiClient.GetAccountBalance(ctx, model.AccountBalanceParams{MarginCoin: model.ParseMarginCoin("usdt")}); err != nil {
				status = fmt.Sprintf("Credentials are invalid: %v", err)
			} else {

				if folderPath != "" {
					_, err := os.Stat(folderPath)
					if os.IsNotExist(err) {
						status = fmt.Sprintf("Folder path does not exist: %v", err)
					} else if err != nil {
						status = fmt.Sprintf("Error checking folder path: %v", err)
					} else {
						config.Mtx.Lock()
						config.ApiKey = apiKey
						config.SecretKey = secretKey
						config.ProfitAndLossFile = folderPath
						config.Mtx.Unlock()

						err := SaveConfig(config)
						if err != nil {
							status = fmt.Sprintf("Error saving config: %v", err)
						} else {
							status = "Configuration saved successfully!"

							go func() { config.Changed <- struct{}{} }()
						}
					}
				}
			}
		}

		apiKeyField := ui.NewLabeledInput(th, "API Key:", "Enter API Key", &apiKeyInput)
		secretKeyField := ui.NewLabeledInput(th, "Secret Key:", "Enter Secret Key", &secretKeyInput)
		folderPathField := ui.NewLabeledInput(th, "Folder Path:", "Enter Folder Path", &folderPathInput)

		folderPathWithButton := func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Spacing:   layout.SpaceBetween,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(0.8, func(gtx layout.Context) layout.Dimensions {
					return folderPathField.Layout(gtx)
				}),
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						selectBtn := material.Button(th, &selectFolderBtn, "Select")
						return selectBtn.Layout(gtx)
					})
				}),
			)
		}

		return ui.VerticalLayout(gtx,
			ui.Title(th, "BitUnix Configuration"),
			apiKeyField.Layout,
			secretKeyField.Layout,
			folderPathWithButton,
			ui.CenteredButton(th, &saveButton, "Save Configuration"),
			ui.CenteredButton(th, &closeButton, "Close"),
			ui.StatusText(th, status),
		)
	}

	return ui.RunWindow(w, configHandler, th)
}

func ShowFolderPicker(log *logger.Logger) string {
	var command *exec.Cmd
	var output []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		script := `
Add-Type -AssemblyName System.Windows.Forms
$folderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog
$folderBrowser.Description = "Select a folder"
$folderBrowser.RootFolder = [System.Environment+SpecialFolder]::MyComputer
if ($folderBrowser.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $folderBrowser.SelectedPath
}
`
		command = exec.Command("powershell", "-Command", script)

	default:
		log.Error("Unsupported platform for folder picker:", runtime.GOOS)
		return ""
	}

	output, err = command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {

			return ""
		}
		log.Error("error showing folder picker:", err)
		return ""
	}

	folderPath := strings.TrimSpace(string(output))

	if folderPath == "" {
		return ""
	}

	_, err = os.Stat(folderPath)
	if err != nil {
		log.Error("error validating selected folder:", err)
		return ""
	}

	return folderPath
}
