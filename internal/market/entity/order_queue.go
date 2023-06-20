package entity

type OrderQueue struct {
	Orders []*Order
}

//Less -> vai dizer que o valor i é menor que um valor j
//Swap -> inverte o i com o j
//Len -> ver o tamanho dos dados que eu tenho
//Push -> adiciona novos itens
//Pop -> remove os itens de uma posição

func (oq *OrderQueue) Less(i, j int) bool {
	return oq.Orders[i].Price < oq.Orders[j].Price
}

func (oq *OrderQueue) Swap(i, j int) {
	oq.Orders[i], oq.Orders[j] = oq.Orders[j], oq.Orders[i]
}

func (oq *OrderQueue) Len() int {
	return len(oq.Orders)
}

func (oq *OrderQueue) Push(x interface{}) { //interface{} = tipagem generica, é igual 'any' de ts
	oq.Orders = append(oq.Orders, x.(*Order))
}

func (oq *OrderQueue) Pop() interface{} {
	oldOrdersValue := oq.Orders
	ordersQuantity := len(oldOrdersValue)
	item := oldOrdersValue[ordersQuantity-1]
	oq.Orders = oldOrdersValue[0 : ordersQuantity-1]
	return item
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{}
}
