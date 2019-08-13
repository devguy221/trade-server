package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CreateOrderEndByte = byte('c')
	FillOrderEndByte   = byte('f')
	CancelOrderEndByte = byte('d')
	Minute             = byte(0x10)
	Hour               = byte(0x20)
	Day                = byte(0x30)
)

type CandleStick struct {
	OpenPrice      sdk.Dec `json:"open"`
	ClosePrice     sdk.Dec `json:"close"`
	HighPrice      sdk.Dec `json:"high"`
	LowPrice       sdk.Dec `json:"low"`
	TotalDeal      sdk.Int `json:"total"`
	EndingUnixTime int64   `json:"unix_time"`
	TimeSpan       byte    `json:"time_span"`
	Market         string  `json:"market"`
}

type Ticker struct {
	Market            string  `json:"market"`
	NewPrice          sdk.Dec `json:"new"`
	OldPriceOneDayAgo sdk.Dec `json:"old"`
}

type PricePoint struct {
	Price  sdk.Dec `json:"p"`
	Amount sdk.Int `json:"a"`
}

type Subscriber interface {
	Detail() interface{}
}

type SubscribeManager interface {
	GetSlashSubscribeInfo() []Subscriber
	GetHeightSubscribeInfo() []Subscriber

	//The returned subscribers have detailed information of markets
	//one subscriber can subscribe tickers from no more than 100 markets
	GetTickerSubscribeInfo() []Subscriber

	//The returned subscribers have detailed information of timespan
	GetCandleStickSubscribeInfo() map[string][]Subscriber

	//the map keys are markets' names
	GetDepthSubscribeInfo() map[string][]Subscriber
	GetDealSubscribeInfo() map[string][]Subscriber

	//the map keys are bancor contracts' names
	GetBancorInfoSubscribeInfo() map[string][]Subscriber

	//the map keys are tokens' names
	GetCommentSubscribeInfo() map[string][]Subscriber

	//the map keys are accounts' bech32 addresses
	GetOrderSubscribeInfo() map[string][]Subscriber
	GetBancorTradeSubscribeInfo() map[string][]Subscriber
	GetIncomeSubscribeInfo() map[string][]Subscriber
	GetUnbondingSubscribeInfo() map[string][]Subscriber
	GetRedelegationSubscribeInfo() map[string][]Subscriber
	GetUnlockSubscribeInfo() map[string][]Subscriber
	GetTxSubscribeInfo() map[string][]Subscriber

	PushSlash(subscriber Subscriber, info []byte)
	PushHeight(subscriber Subscriber, info []byte)
	PushTicker(subscriber Subscriber, t []*Ticker)
	PushDepthSell(subscriber Subscriber, delta map[*PricePoint]bool)
	PushDepthBuy(subscriber Subscriber, delta map[*PricePoint]bool)
	PushCandleStick(subscriber Subscriber, cs *CandleStick)
	PushDeal(subscriber Subscriber, info []byte)
	PushCreateOrder(subscriber Subscriber, info []byte)
	PushFillOrder(subscriber Subscriber, info []byte)
	PushCancelOrder(subscriber Subscriber, info []byte)
	PushBancorInfo(subscriber Subscriber, info []byte)
	PushBancorTrade(subscriber Subscriber, info []byte)
	PushIncome(subscriber Subscriber, info []byte)
	PushUnbonding(subscriber Subscriber, info []byte)
	PushRedelegation(subscriber Subscriber, info []byte)
	PushUnlock(subscriber Subscriber, info []byte)
	PushTx(subscriber Subscriber, info []byte)
	PushComment(subscriber Subscriber, info []byte)
}

type Consumer interface {
	ConsumeMessage(msgType string, bz []byte)
}

type Querier interface {
	// All these functions can be safely called in goroutines.
	QueryTikers(marketList []string) []*Ticker
	QueryBlockTime(height int64, count int) []int64
	QueryDepth(market string, count int) (sell []*PricePoint, buy []*PricePoint)
	QueryCandleStick(market string, timespan byte, time int64, sid int64, count int) [][]byte

	QueryOrder(account string, time int64, sid int64, count int) (data [][]byte, tags []byte, timesid []int64)

	QueryDeal(market string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryBancorInfo(market string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryBancorTrade(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryRedelegation(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryUnbonding(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryUnlock(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryIncome(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryTx(account string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QueryComment(token string, time int64, sid int64, count int) (data [][]byte, timesid []int64)
	QuerySlash(time int64, sid int64, count int) (data [][]byte, timesid []int64)
}