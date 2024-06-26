package gofer

import (
	"encoding/json"
	"fmt"

	"github.com/toknowwhy/theunit-oracle/internal/query"
	pkgEthereum "github.com/toknowwhy/theunit-oracle/pkg/ethereum"
	"github.com/toknowwhy/theunit-oracle/pkg/gofer/origins"
)

func parseParamsSymbolAliases(params json.RawMessage) (origins.SymbolAliases, error) {
	if params == nil {
		return nil, fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		SymbolAliases origins.SymbolAliases `json:"symbolAliases"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.SymbolAliases, nil
}

func parseParamsAPIKey(params json.RawMessage) (string, error) {
	if params == nil {
		return "", fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		APIKey string `json:"apiKey"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return "", fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.APIKey, nil
}

func parseParamsContracts(params json.RawMessage) (origins.ContractAddresses, error) {
	if params == nil {
		return nil, fmt.Errorf("invalid origin parameters")
	}

	var res struct {
		Contracts origins.ContractAddresses `json:"contracts"`
	}
	err := json.Unmarshal(params, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin symbol aliases from params: %w", err)
	}
	return res.Contracts, nil
}

//nolint:funlen,gocyclo
func NewHandler(
	origin string,
	wp query.WorkerPool,
	cli pkgEthereum.Client,
	params json.RawMessage) (origins.Handler, error) {

	aliases, err := parseParamsSymbolAliases(params)
	if err != nil {
		return nil, err
	}
	switch origin {
	case "balancer":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Balancer{
			WorkerPool:        wp,
			ContractAddresses: contracts,
		}, aliases), nil
	case "binance":
		return origins.NewBaseExchangeHandler(origins.Binance{WorkerPool: wp}, aliases), nil
	case "bitfinex":
		return origins.NewBaseExchangeHandler(origins.Bitfinex{WorkerPool: wp}, aliases), nil
	case "bitstamp":
		return origins.NewBaseExchangeHandler(origins.Bitstamp{WorkerPool: wp}, aliases), nil
	case "bitthumb", "bithumb":
		return origins.NewBaseExchangeHandler(origins.BitThump{WorkerPool: wp}, aliases), nil
	case "bittrex":
		return origins.NewBaseExchangeHandler(origins.Bittrex{WorkerPool: wp}, aliases), nil
	case "coinbase", "coinbasepro":
		return origins.NewBaseExchangeHandler(origins.CoinbasePro{WorkerPool: wp}, aliases), nil
	case "cryptocompare":
		return origins.NewBaseExchangeHandler(origins.CryptoCompare{WorkerPool: wp}, aliases), nil
	case "coinmarketcap":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.CoinMarketCap{WorkerPool: wp, APIKey: apiKey},
			aliases,
		), nil
	case "ddex":
		return origins.NewBaseExchangeHandler(origins.Ddex{WorkerPool: wp}, aliases), nil
	case "folgory":
		return origins.NewBaseExchangeHandler(origins.Folgory{WorkerPool: wp}, aliases), nil
	case "ftx":
		return origins.NewBaseExchangeHandler(origins.Ftx{WorkerPool: wp}, aliases), nil
	case "fx":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.Fx{WorkerPool: wp, APIKey: apiKey},
			aliases,
		), nil
	case "gateio":
		return origins.NewBaseExchangeHandler(origins.Gateio{WorkerPool: wp}, aliases), nil
	case "gemini":
		return origins.NewBaseExchangeHandler(origins.Gemini{WorkerPool: wp}, aliases), nil
	case "hitbtc":
		return origins.NewBaseExchangeHandler(origins.Hitbtc{WorkerPool: wp}, aliases), nil
	case "huobi":
		return origins.NewBaseExchangeHandler(origins.Huobi{WorkerPool: wp}, aliases), nil
	case "kraken":
		return origins.NewBaseExchangeHandler(origins.Kraken{WorkerPool: wp}, aliases), nil
	case "kucoin":
		return origins.NewBaseExchangeHandler(origins.Kucoin{WorkerPool: wp}, aliases), nil
	case "kyber":
		return origins.NewBaseExchangeHandler(origins.Kyber{WorkerPool: wp}, aliases), nil
	case "loopring":
		return origins.NewBaseExchangeHandler(origins.Loopring{WorkerPool: wp}, aliases), nil
	case "okex":
		return origins.NewBaseExchangeHandler(origins.Okex{WorkerPool: wp}, aliases), nil
	case "openexchangerates":
		apiKey, err := parseParamsAPIKey(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(
			origins.OpenExchangeRates{WorkerPool: wp, APIKey: apiKey},
			aliases,
		), nil
	case "poloniex":
		return origins.NewBaseExchangeHandler(origins.Poloniex{WorkerPool: wp}, aliases), nil
	case "sushiswap":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Sushiswap{
			WorkerPool:        wp,
			ContractAddresses: contracts,
		}, aliases), nil
	case "curve", "curvefinance":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewCurveFinance(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "balancerV2":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewBalancerV2(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "wsteth":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		h, err := origins.NewWrappedStakedETH(cli, contracts)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(*h, aliases), nil
	case "uniswap", "uniswapV2":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.Uniswap{
			WorkerPool:        wp,
			ContractAddresses: contracts,
		}, aliases), nil
	case "uniswapV3":
		contracts, err := parseParamsContracts(params)
		if err != nil {
			return nil, err
		}
		return origins.NewBaseExchangeHandler(origins.UniswapV3{
			WorkerPool:        wp,
			ContractAddresses: contracts,
		}, aliases), nil
	case "upbit":
		return origins.NewBaseExchangeHandler(origins.Upbit{WorkerPool: wp}, aliases), nil
	}

	return nil, origins.ErrUnknownOrigin
}
