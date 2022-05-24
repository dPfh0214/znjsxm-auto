package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type position struct {
	Symbol          string  `json:"symbol"`
	EntryPrice      float64 `json:"entryPrice"`
	MarkPrice       float64 `json:"markPrice"`
	Pnl             float64 `json:"pnl"`
	Roe             float64 `json:"roe"`
	Amount          float64 `json:"amount"`
	UpdateTimeStamp int     `json:"updateTimeStamp"`
	Yellow          bool    `json:"yellow"`
	TradeBefore     bool    `json:"tradeBefore"`
}

var defaultClient = &http.Client{Timeout: 10 * time.Second}
var positions []position
var bf = new(binanceFunction)

var ratio = big.NewFloat(1000)

func main() {
	setEnv()

	err := getJson("https://laplataquant.me", &positions)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("start")
	startBot()
}

func startBot() {
	for {
		var tempPositions []position
		err := getJson("https://laplataquant.me", &tempPositions)
		if err != nil {
			log.Println(err)
			continue
		}
		checkDiff(tempPositions)
		time.Sleep(time.Millisecond * 500)
	}
}

func checkDiff(tempPositions []position) {
	if len(positions) > 0 {
		for _, val := range positions {
			amt := new(big.Float).SetFloat64(val.Amount)
			sellAll := true
			for _, val2 := range tempPositions {
				if val.Symbol == val2.Symbol {
					sellAll = false

					amt2 := new(big.Float).SetFloat64(val2.Amount)

					if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
						if amt2.Cmp(new(big.Float).SetFloat64(0)) == 1 {
							if amt.Cmp(amt2) == -1 {
								log.Println(val.Symbol + " changed LONG++ ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt2, amt)

								bf.createFutureLongOrder(val.Symbol, makeQuantity(difference), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))

							} else if amt.Cmp(amt2) == 1 {
								log.Println(val.Symbol + " changed LONG-- ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt, amt2)

								bf.closePosition(val.Symbol, makeQuantity(difference), futures.PositionSideTypeLong, makeLeverage(val))
							}

						} else {
							log.Println(val.Symbol + " sellAll LONG ==================================")
							printRes(val)

							bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeLong, makeLeverage(val))

							log.Println(val.Symbol + " created SHORT ==================================")
							printRes(val2)

							bf.createFutureShortOrder(val.Symbol, makeQuantity(amt2), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))
						}
					} else {
						if amt2.Cmp(new(big.Float).SetFloat64(0)) == -1 {
							if amt.Cmp(amt2) == 1 {
								log.Println(val.Symbol + " changed SHORT++ ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt2.Abs(amt2), amt.Abs(amt))

								bf.createFutureShortOrder(val.Symbol, makeQuantity(difference), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))

							} else if amt.Cmp(amt2) == -1 {
								log.Println(val.Symbol + " changed SHORT-- ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt.Abs(amt), amt2.Abs(amt2))

								bf.closePosition(val.Symbol, makeQuantity(difference), futures.PositionSideTypeLong, makeLeverage(val))
							}

						} else {
							log.Println(val.Symbol + " sellAll SHORT ==================================")
							printRes(val)

							bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeShort, makeLeverage(val))

							log.Println(val.Symbol + " created LONG ==================================")
							printRes(val2)

							bf.createFutureLongOrder(val.Symbol, makeQuantity(amt2), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))
						}
					}
				}
			}
			if sellAll {
				if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
					log.Println(val.Symbol + " sellAll LONG ==================================")
					printRes(val)

					bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeLong, makeLeverage(val))
				} else {
					log.Println(val.Symbol + " sellAll SHORT ==================================")
					printRes(val)

					bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeShort, makeLeverage(val))
				}
			}

		}
		for _, val := range tempPositions {
			amt := new(big.Float).SetFloat64(val.Amount)
			exist := false
			for _, val2 := range positions {
				if val.Symbol == val2.Symbol {
					exist = true
				}
			}
			if !exist {
				if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
					log.Println(val.Symbol + " created LONG ==================================")
					printRes(val)

					bf.createFutureLongOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))

				} else {
					log.Println(val.Symbol + " created SHORT ==================================")
					printRes(val)

					bf.createFutureShortOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))
				}
			}
		}
	} else {
		for _, val := range tempPositions {
			amt := new(big.Float).SetFloat64(val.Amount)

			if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
				log.Println(val.Symbol + " created LONG ==================================")
				printRes(val)

				bf.createFutureLongOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))

			} else {
				log.Println(val.Symbol + " created SHORT ==================================")
				printRes(val)

				bf.createFutureShortOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", makeLeverage(val))
			}
		}
	}

	positions = tempPositions
}

func makeRecord() {
	// res, err := db.Exec("INSERT INTO records(symbol,entryprice,marketprice,pnl,roe,amount,update_timestamp,yellow,tradebefore) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)", "BTCUSTD", 40501.38888889, 40573.11757737, 573.8295078, 0.01767887, 8.0, 1650599004533, true, false)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println(res)
}

func getJson(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	res, err := defaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &target)
}

func printRes(position position) {
	log.Printf("symbol: %v\nentryPrice: %v\nmarkprice: %v\npnl: %v\nroe: %v\namount: %v\nupdateTimeStamp: %v\nyellow: %v\ntradeBefore: %v\n", position.Symbol, position.EntryPrice, position.MarkPrice, position.Pnl, position.Roe, position.Amount, position.UpdateTimeStamp, position.Yellow, position.TradeBefore)
}

func makeQuantity(amount *big.Float) string {
	return fmt.Sprintf("%.3f", new(big.Float).Quo(amount, ratio))
}

func makeLeverage(p position) (leverage int) {
	ep := new(big.Float).SetFloat64(p.EntryPrice)
	a := new(big.Float).SetFloat64(p.Amount)
	roe := new(big.Float).SetFloat64(p.Roe)
	pnl := new(big.Float).SetFloat64(p.Pnl)
	leverageStr := fmt.Sprintf("%.0f", new(big.Float).Quo(new(big.Float).Mul(ep, a.Abs(a)), new(big.Float).Mul(new(big.Float).Quo(new(big.Float).SetInt(big.NewInt(1)), roe), pnl)))
	leverage, _ = strconv.Atoi(leverageStr)
	if leverage > 20 {
		leverage = 20
	}
	return leverage
}

func setEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
	apiKey = os.Getenv("API_KEY")
	secretKey = os.Getenv("SECRET_KEY")
	dbHost = os.Getenv("DB_HOST")
	dbPort, _ = strconv.Atoi(os.Getenv("DB_PORT"))
	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbName = os.Getenv("DB_NAME")

	client = binance.NewClient(apiKey, secretKey)
	futuresClient = binance.NewFuturesClient(apiKey, secretKey)   // USDT-M Futures
	deliveryClient = binance.NewDeliveryClient(apiKey, secretKey) // Coin-M Futures
}
