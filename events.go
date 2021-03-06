package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

const (
	ITEM_PRICE            = "PRICE"
	GREATER_THAN          = ">"
	GREATER_THAN_OR_EQUAL = ">="
	LESS_THAN             = "<"
	LESS_THAN_OR_EQUAL    = "<="
	IS_EQUAL              = "=="
	ACTION_SMS_NOTIFY     = "SMS"
	ACTION_CONSOLE_PRINT  = "CONSOLE_PRINT"
)

var (
	ErrInvalidItem         = errors.New("Invalid item.")
	ErrInvalidCondition    = errors.New("Invalid conditional option.")
	ErrInvalidAction       = errors.New("Invalid action.")
	ErrExchangeDisabled    = errors.New("Desired exchange is disabled.")
	ErrFiatCurrencyInvalid = errors.New("Invalid fiat currency.")
)

type Event struct {
	ID             int
	Exchange       string
	Item           string
	Condition      string
	FirstCurrency  string
	SecondCurrency string
	Action         string
	Executed       bool
}

var Events []*Event

func AddEvent(Exchange, Item, Condition, FirstCurrency, SecondCurrency, Action string) (int, error) {
	err := IsValidEvent(Exchange, Item, Condition, Action)

	if err != nil {
		return 0, err
	}

	Event := &Event{}

	if len(Events) == 0 {
		Event.ID = 0
	} else {
		Event.ID = len(Events) + 1
	}

	Event.Exchange = Exchange
	Event.Item = Item
	Event.Condition = Condition
	Event.FirstCurrency = FirstCurrency
	Event.SecondCurrency = SecondCurrency
	Event.Action = Action
	Event.Executed = false
	Events = append(Events, Event)
	return Event.ID, nil
}

func RemoveEvent(EventID int) bool {
	for i, x := range Events {
		if x.ID == EventID {
			Events = append(Events[:i], Events[i+1:]...)
			return true
		}
	}
	return false
}

func GetEventCounter() (int, int) {
	total := len(Events)
	executed := 0

	for _, x := range Events {
		if x.Executed {
			executed++
		}
	}
	return total, executed
}

func (e *Event) ExecuteAction() bool {
	if StringContains(e.Action, ",") {
		action := SplitStrings(e.Action, ",")
		if action[0] == ACTION_SMS_NOTIFY {
			message := fmt.Sprintf("Event triggered: %s", e.EventToString())
			if action[1] == "ALL" {
				SMSSendToAll(message)
			} else {
				SMSNotify(SMSGetNumberByName(action[1]), message)
			}
		}
	} else {
		log.Printf("Event triggered: %s", e.EventToString())
	}
	return true
}

func (e *Event) EventToString() string {
	condition := SplitStrings(e.Condition, ",")
	return fmt.Sprintf("If the %s%s %s on %s is %s then %s.", e.FirstCurrency, e.SecondCurrency, e.Item, e.Exchange, condition[0]+" "+condition[1], e.Action)
}

func (e *Event) CheckCondition() bool {
	lastPrice := 0.00
	condition := SplitStrings(e.Condition, ",")
	targetPrice, _ := strconv.ParseFloat(condition[1], 64)

	ticker, err := GetTickerByExchange(e.Exchange)
	if err != nil {
		return false
	}

	lastPrice = ticker.Price[e.FirstCurrency][e.SecondCurrency].Last

	if lastPrice == 0 {
		return false
	}

	switch condition[0] {
	case GREATER_THAN:
		{
			if lastPrice > targetPrice {
				return e.ExecuteAction()
			}
		}
	case GREATER_THAN_OR_EQUAL:
		{
			if lastPrice >= targetPrice {
				return e.ExecuteAction()
			}
		}
	case LESS_THAN:
		{
			if lastPrice < targetPrice {
				return e.ExecuteAction()
			}
		}
	case LESS_THAN_OR_EQUAL:
		{
			if lastPrice <= targetPrice {
				return e.ExecuteAction()
			}
		}
	case IS_EQUAL:
		{
			if lastPrice == targetPrice {
				return e.ExecuteAction()
			}
		}
	}
	return false
}

func IsValidEvent(Exchange, Item, Condition, Action string) error {
	if !IsValidExchange(Exchange) {
		return ErrExchangeDisabled
	}

	if !IsValidItem(Item) {
		return ErrInvalidItem
	}

	if !StringContains(Condition, ",") {
		return ErrInvalidCondition
	}

	condition := SplitStrings(Condition, ",")

	if !IsValidCondition(condition[0]) || len(condition[1]) == 0 {
		return ErrInvalidCondition
	}

	if StringContains(Action, ",") {
		action := SplitStrings(Action, ",")

		if action[0] != ACTION_SMS_NOTIFY {
			return ErrInvalidAction
		}

		if action[1] != "ALL" && SMSGetNumberByName(action[1]) == ErrSMSContactNotFound {
			return ErrInvalidAction
		}
	} else {
		if Action != ACTION_CONSOLE_PRINT {
			return ErrInvalidAction
		}
	}
	return nil
}

func CheckEvents() {
	for {
		total, executed := GetEventCounter()
		if total > 0 && executed != total {
			for _, event := range Events {
				if !event.Executed {
					success := event.CheckCondition()
					if success {
						log.Printf("Event %d triggered on %s successfully.\n", event.ID, event.Exchange)
						event.Executed = true
					}
				}
			}
		}
	}
}

func IsValidExchange(Exchange string) bool {
	for x, _ := range bot.exchanges {
		if bot.exchanges[x].GetName() == Exchange && bot.exchanges[x].IsEnabled() {
			return true
		}
	}
	return false
}

func IsValidCondition(Condition string) bool {
	switch Condition {
	case GREATER_THAN, GREATER_THAN_OR_EQUAL, LESS_THAN, LESS_THAN_OR_EQUAL, IS_EQUAL:
		return true
	}
	return false
}

func IsValidAction(Action string) bool {
	switch Action {
	case ACTION_SMS_NOTIFY, ACTION_CONSOLE_PRINT:
		return true
	}
	return false
}

func IsValidItem(Item string) bool {
	switch Item {
	case ITEM_PRICE:
		return true
	}
	return false
}
