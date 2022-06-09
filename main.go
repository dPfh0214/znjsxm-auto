package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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
	Leverage        int
	Profit          float64
}

var defaultClient = &http.Client{Timeout: 10 * time.Second}
var positions []position
var bf = new(binanceFunction)

var ratio = big.NewFloat(200)
var leverage = 15

func main() {
	setEnv()
	setDb()

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

		// tempPositions = append(tempPositions, position{Symbol: "BTCUSDT", EntryPrice: 28925.78720123, MarkPrice: 29513.16189206, Pnl: 3524.2481448, Roe: 5.9706369, Amount: 6, UpdateTimeStamp: 1653590087394, Yellow: true, TradeBefore: false})
		checkDiff(tempPositions)
		time.Sleep(time.Millisecond * 1000)
	}
}

func checkDiff(tempPositions []position) {
	if len(positions) > 0 {
		for _, val := range positions {
			leverage, _ := makeLeverage(val)
			if leverage == -1 {
				return
			}
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
								makeRecord(val, "open", "long")

								bf.createFutureLongOrder(val.Symbol, makeQuantity(difference), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)

							} else if amt.Cmp(amt2) == 1 {
								log.Println(val.Symbol + " changed LONG-- ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt, amt2)
								pnl := new(big.Float).SetFloat64(val2.Pnl)
								val.Profit, _ = strconv.ParseFloat(makeQuantity(new(big.Float).Mul(new(big.Float).Quo(pnl.Abs(pnl), new(big.Float).Mul(new(big.Float).SetFloat64(val2.Roe), new(big.Float).SetFloat64(100))), difference)), 64)
								makeRecord(val, "close", "long")

								bf.closePosition(val.Symbol, makeQuantity(difference), futures.PositionSideTypeLong, leverage)
							}

						} else {
							log.Println(val.Symbol + " sellAll LONG ==================================")
							printRes(val)
							val.Profit = val.Pnl
							makeRecord(val, "close", "long")

							bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeLong, leverage)

							log.Println(val.Symbol + " created SHORT ==================================")
							printRes(val2)
							makeRecord(val, "open", "short")

							bf.createFutureShortOrder(val.Symbol, makeQuantity(amt2), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)
						}
					} else {
						if amt2.Cmp(new(big.Float).SetFloat64(0)) == -1 {
							if amt.Cmp(amt2) == 1 {
								log.Println(val.Symbol + " changed SHORT++ ==================================")
								printRes(val2)
								makeRecord(val, "open", "short")

								difference := new(big.Float).Sub(amt2.Abs(amt2), amt.Abs(amt))

								bf.createFutureShortOrder(val.Symbol, makeQuantity(difference), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)

							} else if amt.Cmp(amt2) == -1 {
								log.Println(val.Symbol + " changed SHORT-- ==================================")
								printRes(val2)

								difference := new(big.Float).Sub(amt.Abs(amt), amt2.Abs(amt2))
								pnl := new(big.Float).SetFloat64(val2.Pnl)
								val.Profit, _ = strconv.ParseFloat(makeQuantity(new(big.Float).Mul(new(big.Float).Quo(pnl.Abs(pnl), new(big.Float).Mul(new(big.Float).SetFloat64(val2.Roe), new(big.Float).SetFloat64(100))), difference)), 64)

								makeRecord(val, "close", "short")

								bf.closePosition(val.Symbol, makeQuantity(difference), futures.PositionSideTypeLong, leverage)
							}

						} else {
							log.Println(val.Symbol + " sellAll SHORT ==================================")
							printRes(val)
							val.Profit = val.Pnl
							makeRecord(val, "close", "short")

							bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeShort, leverage)

							log.Println(val.Symbol + " created LONG ==================================")
							printRes(val2)
							makeRecord(val, "open", "long")

							bf.createFutureLongOrder(val.Symbol, makeQuantity(amt2), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)
						}
					}
				}
			}
			if sellAll {
				if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
					log.Println(val.Symbol + " sellAll LONG ==================================")
					printRes(val)
					val.Profit = val.Pnl
					makeRecord(val, "close", "long")

					bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeLong, leverage)
				} else {
					log.Println(val.Symbol + " sellAll SHORT ==================================")
					printRes(val)
					val.Profit = val.Pnl
					makeRecord(val, "close", "short")

					bf.closePosition(val.Symbol, makeQuantity(amt), futures.PositionSideTypeShort, leverage)
				}
			}

		}
		for _, val := range tempPositions {
			leverage, _ := makeLeverage(val)
			if leverage == -1 {
				return
			}
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
					makeRecord(val, "open", "long")

					bf.createFutureLongOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)

				} else {
					log.Println(val.Symbol + " created SHORT ==================================")
					printRes(val)
					makeRecord(val, "open", "short")

					bf.createFutureShortOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)
				}
			}
		}
	} else {
		for _, val := range tempPositions {
			leverage, _ := makeLeverage(val)
			if leverage == -1 {
				return
			}
			amt := new(big.Float).SetFloat64(val.Amount)

			if amt.Cmp(new(big.Float).SetFloat64(0)) == 1 {
				log.Println(val.Symbol + " created LONG ==================================")
				printRes(val)
				makeRecord(val, "open", "long")

				bf.createFutureLongOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)

			} else {
				log.Println(val.Symbol + " created SHORT ==================================")
				printRes(val)
				makeRecord(val, "open", "short")

				bf.createFutureShortOrder(val.Symbol, makeQuantity(amt), fmt.Sprintf("%f", val.MarkPrice), "market", leverage)
			}
		}
	}

	positions = tempPositions
}

func makeRecord(p position, t string, pt string) {
	_, leverage := makeLeverage(p)
	a := p.Amount
	res, err := db.Exec("INSERT INTO records(symbol,entryprice,marketprice,pnl,roe,amount,update_timestamp,yellow,tradebefore,type,position_type,leverage,profit) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)", p.Symbol, p.EntryPrice, p.MarkPrice, p.Pnl, p.Roe, math.Abs(a), p.UpdateTimeStamp, p.Yellow, p.TradeBefore, t, pt, leverage, p.Profit)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res)
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

func printRes(p position) {
	_, leverage := makeLeverage(p)
	log.Printf("symbol: %v\nentryPrice: %v\nmarkprice: %v\npnl: %v\nroe: %v\namount: %v\nupdateTimeStamp: %v\nyellow: %v\ntradeBefore: %v\nleverage: %v\n", p.Symbol, p.EntryPrice, p.MarkPrice, p.Pnl, p.Roe, p.Amount, p.UpdateTimeStamp, p.Yellow, p.TradeBefore, leverage)
}

func makeQuantity(amount *big.Float) string {
	return fmt.Sprintf("%.3f", new(big.Float).Quo(amount.Abs(amount), ratio))
}

func makeLeverage(p position) (leverage int, quantLeverage int) {
	if p.Roe == 0 || p.Pnl == 0 {
		return -1, -1
	}
	ep := new(big.Float).SetFloat64(p.EntryPrice)
	a := new(big.Float).SetFloat64(p.Amount)
	roe := new(big.Float).SetFloat64(p.Roe)
	pnl := new(big.Float).SetFloat64(p.Pnl)

	leverageStr := fmt.Sprintf("%.0f", new(big.Float).Quo(new(big.Float).Mul(ep, a.Abs(a)), new(big.Float).Mul(new(big.Float).Quo(new(big.Float).SetInt(big.NewInt(1)), roe), pnl)))
	leverage, _ = strconv.Atoi(leverageStr)
	quantLeverage = leverage
	if leverage > 20 {
		leverage = 20
	}
	return leverage, quantLeverage
}

func setEnv() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
