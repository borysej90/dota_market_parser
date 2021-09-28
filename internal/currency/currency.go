package currency

// Currency stores currencies in ISO 4217 format by their names.
var Currency = map[string]int{
	"UAH": 980,
	"USD": 840,
	"RUB": 643,
	"EUR": 978,
}

type Parser interface {
	GetCurrencyRate(currencyName string) (float32, error)
}
