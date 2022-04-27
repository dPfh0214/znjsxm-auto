package main

import (
	"context"
	"log"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/delivery"
	"github.com/adshao/go-binance/v2/futures"
)

var (
	apiKey    = ""
	secretKey = ""
)

var client *binance.Client
var futuresClient *futures.Client   // USDT-M Futures
var deliveryClient *delivery.Client // Coin-M Futures

type binanceFunction struct{}

func (*binanceFunction) createBuyOrder(symbol string, quantity string, price string) {
	log.Println("createBuyOrder:")
	order, err := client.NewCreateOrderService().Symbol(symbol).
		Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(quantity).
		Price(price).Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(order)
}

func (*binanceFunction) createSellOrder(symbol string, quantity string, price string) {
	log.Println("createSellOrder:")
	order, err := client.NewCreateOrderService().Symbol(symbol).
		Side(binance.SideTypeSell).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(quantity).
		Price(price).Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(order)
}

func (*binanceFunction) cancelOrder(symbol string, orderId int64) {
	_, err := client.NewCancelOrderService().Symbol(symbol).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
}

func (*binanceFunction) getOrder(symbol string, orderId int64) {
	log.Println("getOrder:")
	order, err := client.NewGetOrderService().Symbol(symbol).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(order)
}

func (*binanceFunction) getOpenOrderList(symbol string) {
	log.Println("getOpenOrderList:")
	openOrders, err := client.NewListOpenOrdersService().Symbol(symbol).
		Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	for _, o := range openOrders {
		log.Println(o)
	}

}

func (*binanceFunction) getOrderList(symbol string) {
	log.Println("getOrderList:")
	orders, err := client.NewListOrdersService().Symbol(symbol).
		Do(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
	for _, o := range orders {
		log.Println(o)
	}
}

func (*binanceFunction) setLeverage(symbol string) {
	// log.Println("setLeverage:")
	_, err := futuresClient.NewChangeLeverageService().Leverage(5).Symbol(symbol).Do(context.Background())
	if err != nil {
		// log.Println(err)
		return
	}
	// log.Println(leverage)
}

func (*binanceFunction) setMarginType(symbol string) {
	// log.Println("setMarginType:")
	err := futuresClient.NewChangeMarginTypeService().MarginType(futures.MarginTypeIsolated).Symbol(symbol).Do(context.Background())
	if err != nil {
		// log.Println(err)
		return
	}
}

func (*binanceFunction) setPositionMode() {
	// log.Println("setPositionMode:")
	err := futuresClient.NewChangePositionModeService().DualSide(true).Do(context.Background())

	if err != nil {
		// log.Println(err)
		return
	}
}

func (*binanceFunction) createFutureLongOrder(symbol string, quantity string, price string, orderType string) {
	bf.setLeverage(symbol)
	bf.setMarginType(symbol)
	bf.setPositionMode()

	log.Println("createFutureLongOrder:")
	log.Println(symbol, quantity, price, orderType)

	var err error
	var order *futures.CreateOrderResponse

	switch orderType {
	case "limit":
		order, err = futuresClient.NewCreateOrderService().Symbol(symbol).
			Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(quantity).PositionSide(futures.PositionSideTypeLong).Price(price).
			Do(context.Background())

	case "market":
		order, err = futuresClient.NewCreateOrderService().Symbol(symbol).
			Side(futures.SideTypeBuy).Type(futures.OrderTypeMarket).
			Quantity(quantity).PositionSide(futures.PositionSideTypeLong).
			Do(context.Background())
	}

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(order)
}

func (*binanceFunction) createFutureShortOrder(symbol string, quantity string, price string, orderType string) {
	bf.setLeverage(symbol)
	bf.setMarginType(symbol)
	bf.setPositionMode()

	log.Println("createFutureShortOrder:")
	log.Println(symbol, quantity, price, orderType)
	var err error
	var order *futures.CreateOrderResponse

	switch orderType {
	case "limit":
		order, err = futuresClient.NewCreateOrderService().Symbol(symbol).
			Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(quantity).PositionSide(futures.PositionSideTypeShort).Price(price).
			Do(context.Background())

	case "market":
		order, err = futuresClient.NewCreateOrderService().Symbol(symbol).
			Side(futures.SideTypeBuy).Type(futures.OrderTypeMarket).
			Quantity(quantity).PositionSide(futures.PositionSideTypeShort).
			Do(context.Background())
	}

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(order)
}

func (*binanceFunction) closePosition(symbol string, quantity string, side bool) {
	bf.setLeverage(symbol)
	bf.setMarginType(symbol)
	bf.setPositionMode()

	log.Println("closePosition:")
	log.Println(symbol, quantity, side)

	var positionSide futures.PositionSideType

	if !side {
		positionSide = futures.PositionSideTypeLong
	} else {
		positionSide = futures.PositionSideTypeLong
	}

	order, err := futuresClient.NewCreateOrderService().Symbol(symbol).
		Side(futures.SideTypeSell).Type(futures.OrderTypeMarket).
		Quantity(quantity).PositionSide(positionSide).
		Do(context.Background())

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(order)
}
