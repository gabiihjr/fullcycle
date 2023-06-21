package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Orders        []*Order
	Transactions  []*Transaction
	OrdersChan    chan *Order // input -> recebe ordens do kafka
	OrdersChanOut chan *Order // output
	Wg            *sync.WaitGroup
}

func NewBook(orderChan chan *Order, orderChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Orders:        []*Order{},
		Transactions:  []*Transaction{},
		OrdersChan:    orderChan,
		OrdersChanOut: orderChanOut,
		Wg:            wg,
	}
}

func (book *Book) Trade() {
	// buyOrders := NewOrderQueue()
	buyOrders := make(map[string]*OrderQueue)
	// sellOrders := NewOrderQueue()
	sellOrders := make(map[string]*OrderQueue)

	// heap.Init(buyOrders)
	// heap.Init(sellOrders)

	for order := range book.OrdersChan {
		asset := order.Asset.ID

		if buyOrders[asset] == nil {
			buyOrders[asset] = NewOrderQueue()
			heap.Init(buyOrders[asset])
		}

		if sellOrders[asset] == nil {
			sellOrders[asset] = NewOrderQueue()
			heap.Init(sellOrders[asset])
		}

		if order.OrderType == "BUY" {
			buyOrders[asset].Push(order)
			if sellOrders[asset].Len() > 0 && sellOrders[asset].Orders[0].Price <= order.Price {
				sellOrder := sellOrders[asset].Pop().(*Order)
				if sellOrder.PendingShares > 0 {
					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
					book.AddTransaction(transaction, book.Wg)
					sellOrder.Transactions = append(order.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					book.OrdersChanOut <- sellOrder
					book.OrdersChanOut <- order
					if sellOrder.PendingShares > 0 {
						sellOrders[asset].Push(sellOrder)
					}
				}
			}
		} else if order.OrderType == "SELL" {
			sellOrders[asset].Push(order)
			if buyOrders[asset].Len() > 0 && buyOrders[asset].Orders[0].Price >= order.Price {
				buyOrder := buyOrders[asset].Pop().(*Order)
				if buyOrder.PendingShares > 0 {
					transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
					book.AddTransaction(transaction, book.Wg)
					buyOrder.Transactions = append(buyOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					book.OrdersChanOut <- buyOrder
					book.OrdersChanOut <- order
					if buyOrder.PendingShares > 0 {
						buyOrders[asset].Push(buyOrder)
					}
				}
			}
		}
	}
}

func (book *Book) AddTransaction(transaction *Transaction, wg *sync.WaitGroup) {
	defer wg.Done() //tudo que colocar embaixo será executado e POR ULTIMO o defer é utilizado

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	minShares := sellingShares
	if buyingShares < minShares {
		minShares = buyingShares
	}

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	// transaction.SellingOrder.PendingShares -= minShares
	transaction.AddSellOrderPendingShares(-minShares)
	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.BuyingOrder.Asset.ID, minShares)
	// transaction.BuyingOrder.PendingShares -= minShares
	transaction.AddBuyOrderPendingShares(-minShares)

	// transaction.Total = float64(transaction.Shares) * transaction.BuyingOrder.Price
	transaction.CalculateTotal(transaction.Shares, transaction.BuyingOrder.Price)

	// if transaction.BuyingOrder.PendingShares == 0 {
	// 	transaction.BuyingOrder.Status = "CLOSED"
	// }
	transaction.CloseBuyOrderTransaction()
	// if transaction.SellingOrder.PendingShares == 0 {
	// 	transaction.SellingOrder.Status = "CLOSED"
	// }
	transaction.CloseSellOrderTransaction()
	book.Transactions = append(book.Transactions, transaction)
}
