package user_handler

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/anarakinson/go_stonks/stonks_client/internal/domain"

	pb_market "github.com/anarakinson/go_stonks/stonks_pb/gen/market"
	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

func (h *UserHandler) GetUserID() (string, error) {

	fmt.Print("Enter UserId: ")
	userId, err := h.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nEnd")
			return "", ErrFinish
		}
		return "", ErrStdin
	}
	return strings.ToLower(strings.TrimSpace(userId)), nil
}

func (h *UserHandler) GetMarketID(markets []*pb_market.Market) (string, error) {

	if len(markets) == 0 {
		return "", ErrNoData
	}

	var result string
	var input string
	var err error
	for input != "exit" {
		fmt.Println("Chose market (input number):")
		for i, m := range markets {
			fmt.Printf("\t%d. %s\n", i+1, m.Name)
		}
		input, err = h.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", ErrFinish
			}
			return "", ErrStdin
		}
		input = strings.ToLower(strings.TrimSpace(input))

		marketIdx, err := strconv.Atoi(input)
		marketIdx -= 1
		if err != nil || marketIdx < 0 || marketIdx >= len(markets) {
			fmt.Printf("Number must be positive digit less than %d\n", len(markets))
			continue
		}
		result = markets[marketIdx].Id
		// возвращаем ID
		return result, nil
	}
	// если цикл прервался - прерываем программу
	return "", ErrFinish

}

func (h *UserHandler) GetOrderType() (string, error) {

	var input string
	var err error
	for input != "exit" {
		fmt.Print("Chose type (buy, sell): ")
		input, err = h.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", ErrFinish
			}
			return "", ErrStdin
		}
		input = strings.ToLower(strings.TrimSpace(input))

		if input != "buy" && input != "sell" {
			fmt.Println("Input 'buy' or 'sell'")
			continue
		}
		return input, nil

	}
	// если цикл прервался - прерываем программу
	return "", ErrFinish

}

func (h *UserHandler) GetPrice() (float64, error) {

	var input string
	var err error
	for input != "exit" {
		fmt.Print("Input price (float): ")
		input, err = h.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return 0, ErrFinish
			}
			return 0, ErrStdin
		}
		input = strings.ToLower(strings.TrimSpace(input))

		price, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("Price must be float")
			continue
		}
		return price, nil
	}
	// если цикл прервался - прерываем программу
	return 0, ErrFinish

}

func (h *UserHandler) GetQuantity() (float64, error) {

	var input string
	var err error
	for input != "exit" {
		fmt.Print("Input quantity (float): ")
		input, err = h.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return 0, ErrFinish
			}
			return 0, ErrStdin
		}
		input = strings.ToLower(strings.TrimSpace(input))

		quantity, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("Price must be float")
			continue
		}
		return quantity, nil
	}
	// если цикл прервался - прерываем программу
	return 0, ErrFinish

}

func (h *UserHandler) GetOrder(userID string) (*domain.Order, error) {

	// 1. Выбор рынка
	// получаем список рынков от сервиса
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	markets, err := h.client.GetMarkets(ctx, &pb.GetMarketsRequest{})
	if err != nil {
		return nil, fmt.Errorf("getMarkets failed: %v", err)

	}

	marketID, err := h.GetMarketID(markets.Markets)
	if err != nil {
		return nil, fmt.Errorf("getMarketID failed: %v", err)

	}

	// 2. Тип заказа
	orderType, err := h.GetOrderType()
	if err != nil {
		return nil, fmt.Errorf("getOrderType failed: %v", err)
	}

	// 3. Цена
	price, err := h.GetPrice()
	if err != nil {
		return nil, fmt.Errorf("getPrice failed: %v", err)
	}

	// 4. Количество
	quantity, err := h.GetQuantity()
	if err != nil {
		return nil, fmt.Errorf("getQuantity failed: %v", err)
	}

	return &domain.Order{
		UserID:    userID,
		MarketID:  marketID,
		OrderType: orderType,
		Price:     price,
		Quantity:  quantity,
	}, nil

}
