package main

import (
	"bufio"
	"container/list"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type Price uint // pennies

type Order struct {
	oid   string
	price Price
	size  uint
}

type OrderBookSide struct {
	side        string
	betterPrice func(Price, Price) bool
	ordersList  *list.List
	targetSize  uint
	prevResult  string
}

func NewOrderBookSide(side string, size uint, betterPrice func(Price, Price) bool) *OrderBookSide {
	obs := new(OrderBookSide)
	obs.side = side
	obs.betterPrice = betterPrice
	obs.ordersList = list.New()
	obs.targetSize = size
	obs.prevResult = side + " NA"
	return obs
}

func (obs *OrderBookSide) addOrder(oid string, price Price, size uint) string {
	var curTotal Price
	sizeNeeded := obs.targetSize
	applyOrder := func(o *Order) {
		sizeUsed := min(sizeNeeded, o.size)
		curTotal += o.price * Price(sizeUsed)
		sizeNeeded -= sizeUsed
	}

	inserted := false
	for e := obs.ordersList.Front(); e != nil; e = e.Next() {
		order := e.Value.(*Order)
		if !inserted && obs.betterPrice(price, order.price) {
			// Insert new before current order
			obs.ordersList.InsertBefore(&Order{oid, price, size}, e)
			applyOrder(e.Prev().Value.(*Order))
			inserted = true
		}
		applyOrder(order)

		if inserted && sizeNeeded == 0 {
			break
		}
	}
	if !inserted {
		obs.ordersList.PushBack(&Order{oid, price, size})
		applyOrder(obs.ordersList.Back().Value.(*Order))
	}

	res := obs.side + " NA"
	if sizeNeeded == 0 {
		res = obs.side + fmt.Sprintf(" %d.%02d", curTotal/100, curTotal%100)
	}

	if res != obs.prevResult {
		obs.prevResult = res
		return res
	}
	return ""
}

func (obs *OrderBookSide) reduceOrder(oid string, size uint) string {
	var curTotal Price
	sizeNeeded := obs.targetSize

	found := false
	for e := obs.ordersList.Front(); e != nil; e = e.Next() {
		order := e.Value.(*Order)
		if order.oid == oid {
			if size >= order.size {
				next := e.Next()
				obs.ordersList.Remove(e)
				found = true
				if next == nil {
					break
				}
				e = next
				order = e.Value.(*Order)
			} else {
				order.size -= size
			}
		}
		sizeUsed := min(sizeNeeded, order.size)
		curTotal += order.price * Price(sizeUsed)
		sizeNeeded -= sizeUsed

		if found && sizeNeeded == 0 {
			break
		}
	}

	res := obs.side + " NA"
	if sizeNeeded == 0 {
		res = obs.side + fmt.Sprintf(" %d.%02d", curTotal/100, curTotal%100)
	}

	if res != obs.prevResult {
		obs.prevResult = res
		return res
	}
	return ""
}

func (obs *OrderBookSide) hasOrder(oid string) bool {
	for e := obs.ordersList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Order).oid == oid {
			return true
		}
	}
	return false
}

type Pricer struct {
	bids *OrderBookSide
	asks *OrderBookSide
}

func NewPricer(size uint) *Pricer {
	p := new(Pricer)
	p.bids = NewOrderBookSide("S", size, func(p1 Price, p2 Price) bool { return p1 > p2 })
	p.asks = NewOrderBookSide("B", size, func(p1 Price, p2 Price) bool { return p1 < p2 })
	return p
}

func (p *Pricer) HandleLine(line string) (tm uint, res string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}

	if itm, err := strconv.Atoi(fields[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
		return
	} else {
		tm = uint(itm)
	}

	if len(fields) == 6 && fields[1] == "A" {
		if newTotal := p.handleAdd(line, fields); newTotal != "" {
			res = newTotal
		}
	} else if len(fields) == 4 && fields[1] == "R" {
		if newTotal := p.handleRemove(line, fields); newTotal != "" {
			res = newTotal
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
	}

	return
}

func (p *Pricer) handleAdd(line string, fields []string) string {
	oid := fields[2]
	side := fields[3]

	var price Price
	if fprice, err := strconv.ParseFloat(fields[4], 32); err != nil {
		fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
		return ""
	} else {
		price = Price(math.Round(fprice * 100))
	}

	var size uint
	if isize, err := strconv.Atoi(fields[5]); err != nil {
		fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
		return ""
	} else {
		size = uint(isize)
	}

	if side == "B" {
		return p.bids.addOrder(oid, price, size)
	} else if side == "S" {
		return p.asks.addOrder(oid, price, size)
	}

	fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
	return ""
}

func (p *Pricer) handleRemove(line string, fields []string) string {
	var size uint
	if isize, err := strconv.Atoi(fields[3]); err != nil {
		fmt.Fprintf(os.Stderr, "Unparseable line: %v\n", line)
		return ""
	} else {
		size = uint(isize)
	}

	oid := fields[2]
	if p.bids.hasOrder(oid) {
		return p.bids.reduceOrder(oid, size)
	} else if p.asks.hasOrder(oid) {
		return p.asks.reduceOrder(oid, size)
	}
	// Order not found, ignore
	return ""
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: pricer <size>")
	}
	size, err := strconv.Atoi(os.Args[1])
	if err != nil || size <= 0 {
		log.Fatalf("size argument must be a positive integer")
	}

	pricer := NewPricer(uint(size))

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if tm, res := pricer.HandleLine(line); res != "" {
			fmt.Printf("%v %s\n", tm, res)
		}
	}
}
