package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/tradingiq/bitunix-client/bitunix"
	bitunix_errors "github.com/tradingiq/bitunix-client/errors"
	"github.com/tradingiq/bitunix-client/model"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func runPnl(ctx context.Context, cfg *Config) {
	errChan := make(chan error)

	for {
		ctx, cancel := context.WithCancel(ctx)

		if cfg.SecretKey != "" && cfg.ApiKey != "" {
			err := beeep.Notify("TradingIQ PNL Tracker", "PNL Tracking Started", "assets/information.png")
			if err != nil {
				log.Warning("Could not notify about start of pnl tracking: %v", err)
			}
			go track(ctx, cfg, errChan)
		}

		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := nextMidnight.Sub(now)
		firstTick := time.NewTimer(duration)

		select {
		case err := <-errChan:
			if err != nil {
				log.Error("error while pnl tracking, %v", err)

				cancel()

				switch {
				case errors.Is(err, bitunix_errors.ErrAuthentication), errors.Is(err, bitunix_errors.ErrSignatureError):
					err := beeep.Notify("TradingIQ PNL Tracker", "Authentication failed", "assets/information.png")
					if err != nil {
						log.Warning("Could not notify about authentication error: %v", err)
					}
					mStatus.SetTitle("Authentication Error")
					cfg.SecretKey = ""
					cfg.ApiKey = ""

				case errors.Is(err, bitunix_errors.ErrNetwork), errors.Is(err, bitunix_errors.ErrTimeout):
					err := beeep.Notify("TradingIQ PNL Tracker", "Network connection failed", "assets/information.png")
					if err != nil {
						log.Warning("Could not notify about network error: %v", err)
					}

					mStatus.SetTitle("Timeout Error")

					time.Sleep(5 * time.Minute)
				default:
					mStatus.SetTitle("Error")
				}

			}
		case <-cfg.Changed:
			log.Debug("starting pnl tracking")
			mStatus.SetTitle("Inactive...")

			cancel()
		case <-firstTick.C:
			log.Debug("restarting pnl tracking")
			mStatus.SetTitle("Inactive...")

			cancel()
		case <-ctx.Done():
			log.Debug("exiting pnl tracking")

			mStatus.SetTitle("Exiting...")

			cancel()
			return
		}
	}

}

func track(ctx context.Context, config *Config, errChan chan error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	apiClient, wsClient, err := initClient(ctx, config)
	if err != nil {
		log.Error("failed to create API client: %v", err)
		errChan <- err
		return
	}

	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Error("failed to load timezone: %v", err)
		errChan <- err
		return
	}

	now := time.Now().In(berlin)
	todayMorning := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowMorning := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	realizedPnl, err := fetchInitialBalance(todayMorning, tomorrowMorning, err, apiClient, ctx)
	if err != nil {
		log.Error("failed to fetch initial balance: %v", err)
		errChan <- err
		return
	}

	mStatus.SetTitle(fmt.Sprintf("Running - Todays PnL %.2f", realizedPnl))

	log.Debug("initial balance at application start: %.2f", realizedPnl)
	pnl := NewProfitAndLoss(realizedPnl)

	if config.ProfitAndLossFile != "" {
		if err := SavePnLToFile(realizedPnl, config.ProfitAndLossFile); err != nil {
			log.Warning("failed to save initial PnL to file: %v", err)
		}
	}

	pnl.config = config

	if err := wsClient.SubscribePositions(pnl); err != nil {
		log.Error("failed to subscribe to positions: %v", err)
		errChan <- err
		return
	}

	if err := wsClient.Stream(); err != nil {
		if errors.Is(err, bitunix_errors.ErrConnectionClosed) {
			log.Debug("websocket is ending")

			return
		} else {
			log.Error("failed to stream positions: %v", err.Error())

			errChan <- err
			return
		}
	}
}

func fetchInitialBalance(todayMorning time.Time, tomorrowMorning time.Time, err error, apiClient bitunix.ApiClient, ctx context.Context) (float64, error) {
	params := model.PositionHistoryParams{
		Limit:     100,
		StartTime: &todayMorning,
		EndTime:   &tomorrowMorning,
	}

	posResponse, err := apiClient.GetPositionHistory(ctx, params)
	if err != nil {
		log.Debug("failed to fetch initial positions: %v", err)
		return 0.0, err
	}

	var (
		realizedPnl float64 = 0
	)
	for _, position := range posResponse.Data.Positions {
		realizedPnl += position.RealizedPNL
	}
	return realizedPnl, nil
}

func initClient(ctx context.Context, config *Config) (bitunix.ApiClient, bitunix.PrivateWebsocketClient, error) {
	config.Mtx.Lock()
	defer config.Mtx.Unlock()

	apiClient, err := bitunix.NewApiClient(config.ApiKey, config.SecretKey)
	if err != nil {
		log.Error("failed to create API client: %v", err)
	}

	ws, err := bitunix.NewPrivateWebsocket(ctx, config.ApiKey, config.SecretKey)
	if err != nil {
		log.Error("failed to create WebSocket client: %v", err)
	}
	if err := ws.Connect(); err != nil {
		log.Error("failed to connect to WebSocket client: %v", err)
	}
	return apiClient, ws, err
}

type ProfitAndLoss struct {
	realizedPnl float64
	mtx         sync.Mutex
	config      *Config
}

func NewProfitAndLoss(initialProfitAndLoss float64) *ProfitAndLoss {
	pnl := &ProfitAndLoss{realizedPnl: initialProfitAndLoss, mtx: sync.Mutex{}}

	return pnl
}

func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.Mode().IsRegular()
}

func SavePnLToFile(pnl float64, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	if !IsFile(filePath) {
		filePath = fmt.Sprintf("%s\\pnl.txt", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			log.Error("error creating pnl file:", err)
		}
		file.Close()
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error("error creating pnl file:", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	pnlString := fmt.Sprintf("%.2f", pnl)

	if err := os.WriteFile(filePath, []byte(pnlString), 0644); err != nil {
		return fmt.Errorf("failed to write PnL to file: %w", err)
	}

	log.Debug("saved PnL value %.2f to file: %s", pnl, filePath)
	return nil
}

func (p *ProfitAndLoss) SubscribePosition(message *model.PositionChannelMessage) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	switch message.Data.Event {
	case model.PositionEventClose:
		p.realizedPnl += message.Data.RealizedPNL
		mStatus.SetTitle(fmt.Sprintf("Running - %.2f", p.realizedPnl))
		log.Debug("position close message received, realized pnl is now %.2f", p.realizedPnl)

		if p.config != nil && p.config.ProfitAndLossFile != "" {
			if err := SavePnLToFile(p.realizedPnl, p.config.ProfitAndLossFile); err != nil {
				log.Warning("failed to save updated PnL to file: %v", err)
			}
		}
	}
}
