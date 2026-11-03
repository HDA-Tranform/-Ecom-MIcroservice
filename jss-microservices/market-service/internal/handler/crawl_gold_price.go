package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	"github.com/linhhuynhcoding/jss-microservices/market/config"
	"github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	"go.uber.org/zap"
)

type IGoldPriceCrawler interface {
	Handle(ctx context.Context) error
}

type GoldPriceCrawler struct {
	logger  *zap.Logger
	cfg     config.Config
	queries repository.Store
}

func NewGoldPriceCrawler(ctx context.Context, logger *zap.Logger, cfg config.Config, store repository.Store) IGoldPriceCrawler {
	return &GoldPriceCrawler{
		logger:  logger,
		cfg:     cfg,
		queries: store,
	}
}

// --- Struct để parse JSON từ API ---
type BTMCResponse struct {
	DataList struct {
		Data []map[string]string `json:"Data"`
	} `json:"DataList"`
}

func (c *GoldPriceCrawler) Handle(ctx context.Context) error {
	logger := c.logger.With(zap.String("func", "GoldPriceCrawler"))

	// 1. Call API
	resp, err := http.Get("http://api.btmc.vn/api/BTMCAPI/getpricebtmc?key=3kd8ub1llcg9t45hnoh8hmn7t5kc2v")
	if err != nil {
		logger.Error("Failed to call API", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return err
	}
	logger.Info("Response body", zap.ByteString("body", body))

	// 2. Parse JSON
	var data BTMCResponse
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Error("Failed to parse JSON", zap.Error(err))
		return err
	}

	// 3. Insert/update data
	for _, item := range data.DataList.Data {
		goldID := 1
		goldType := item["@n_1"]
		if goldType == "" {
			// fallback: các row có thể dùng n_2, n_3...
			for k, v := range item {
				if len(k) > 2 && k[:2] == "@n" {
					goldID, err = strconv.Atoi(k[3:])
					if err != nil {
						logger.Error("Failed to parse goldID", zap.Error(err))
						break
					}
					goldType = v
					break
				}
			}
		}

		var pb, ps, d string
		for k, v := range item {
			switch {
			case len(k) > 3 && k[:3] == "@pb":
				pb = v
			case len(k) > 3 && k[:3] == "@ps":
				ps = v
			case len(k) > 2 && k[:2] == "@d":
				d = v
			}
		}

		var buyPrice, sellPrice float64
		fmt.Sscanf(pb, "%f", &buyPrice)
		fmt.Sscanf(ps, "%f", &sellPrice)

		date, err := time.Parse("02/01/2006 15:04", d)
		if err != nil {
			logger.Warn("Invalid date format, skip", zap.String("date", d))
			continue
		}

		g, err := c.queries.UpsertPrice(ctx, repository.UpsertPriceParams{
			Date:      pgtype.Timestamp{Valid: true, Time: date},
			GoldID:    int32(goldID),
			GoldType:  goldType,
			BuyPrice:  utils.ToNumeric(buyPrice),
			SellPrice: utils.ToNumeric(sellPrice),
		})
		if err != nil {
			logger.Error("Failed to upsert gold price", zap.String("goldType", goldType), zap.Error(err))
		} else {
			logger.Info("Upserted gold price", zap.Any("gold_price", g))
		}
	}

	logger.Info("Done updating gold prices")
	return nil
}
