package main

type FireflyAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type FireflyAccountRequest struct {
	Name               string  `json:"name,omitempty"`
	Type               string  `json:"type,omitempty"`
	IBAN               string  `json:"iban,omitempty"`
	BIC                string  `json:"bic,omitempty"`
	AccountNumber      string  `json:"account_number,omitempty"`
	OpeningBalance     string  `json:"opening_balance,omitempty"`
	OpeningBalanceDate string  `json:"opening_balance_date,omitempty"`
	VirtualBalance     string  `json:"virtual_balance,omitempty"`
	CurrencyID         string  `json:"currency_id,omitempty"`
	CurrencyCode       string  `json:"currency_code,omitempty"`
	Active             bool    `json:"active,omitempty"`
	Order              int     `json:"order,omitempty"`
	IncludeNetWorth    bool    `json:"include_net_worth,omitempty"`
	AccountRole        string  `json:"account_role,omitempty"`
	CreditCardType     string  `json:"credit_card_type,omitempty"`
	MonthlyPaymentDate string  `json:"monthly_payment_date,omitempty"`
	LiabilityType      string  `json:"liability_type,omitempty"`
	LiabilityDirection string  `json:"liability_direction,omitempty"`
	Interest           string  `json:"interest,omitempty"`
	InterestPeriod     string  `json:"interest_period,omitempty"`
	Notes              string  `json:"notes,omitempty"`
	Latitude           float64 `json:"latitude,omitempty"`
	Longitude          float64 `json:"longitude,omitempty"`
	ZoomLevel          int     `json:"zoom_level,omitempty"`
}

// RequestParams defines the parameters for the API request.
type RequestParams struct {
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Date  string `json:"date"`
	Type  string `json:"type"`
}

// AccountAttributes defines the attributes of an account.
type AccountAttributes struct {
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
	Active                bool    `json:"active"`
	Order                 int     `json:"order"`
	Name                  string  `json:"name"`
	Type                  string  `json:"type"`
	AccountRole           string  `json:"account_role"`
	CurrencyID            string  `json:"currency_id"`
	CurrencyCode          string  `json:"currency_code"`
	CurrencySymbol        string  `json:"currency_symbol"`
	CurrencyDecimalPlaces int     `json:"currency_decimal_places"`
	CurrentBalance        string  `json:"current_balance"`
	CurrentBalanceDate    string  `json:"current_balance_date"`
	IBAN                  string  `json:"iban"`
	BIC                   string  `json:"bic"`
	AccountNumber         string  `json:"account_number"`
	OpeningBalance        string  `json:"opening_balance"`
	CurrentDebt           string  `json:"current_debt"`
	OpeningBalanceDate    string  `json:"opening_balance_date"`
	VirtualBalance        string  `json:"virtual_balance"`
	IncludeNetWorth       bool    `json:"include_net_worth"`
	CreditCardType        string  `json:"credit_card_type"`
	MonthlyPaymentDate    string  `json:"monthly_payment_date"`
	LiabilityType         string  `json:"liability_type"`
	LiabilityDirection    string  `json:"liability_direction"`
	Interest              string  `json:"interest"`
	InterestPeriod        string  `json:"interest_period"`
	Notes                 string  `json:"notes"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	ZoomLevel             int     `json:"zoom_level"`
}

// AccountData represents a single account entry in the response.
type AccountData struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes AccountAttributes `json:"attributes"`
}

// MetaPagination defines pagination details in the response.
type MetaPagination struct {
	Total       int `json:"total"`
	Count       int `json:"count"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

// Meta contains the metadata of the response.
type Meta struct {
	Pagination MetaPagination `json:"pagination"`
}

type CreateFireflyAccountResponse struct {
	Data AccountData `json:"data"`
}

// FireflyAccountResponse represents the complete API response.
type FireflyAccountResponse struct {
	Data []AccountData `json:"data"`
	Meta Meta          `json:"meta"`
}

type FireflyTransaction struct {
	ErrorIfDuplicateHash bool          `json:"error_if_duplicate_hash"`
	ApplyRules           bool          `json:"apply_rules"`
	FireWebhooks         bool          `json:"fire_webhooks"`
	GroupTitle           string        `json:"group_title"`
	Transactions         []Transaction `json:"transactions"`
}

type Transaction struct {
	Type            string   `json:"type"`
	Date            string   `json:"date"`
	Amount          string   `json:"amount"`
	Description     string   `json:"description"`
	CurrencyID      string   `json:"currency_id"`
	CurrencyCode    string   `json:"currency_code"`
	SourceID        string   `json:"source_id"`
	DestinationID   string   `json:"destination_id"`
	CategoryName    string   `json:"category_name"`
	SourceName      string   `json:"source_name"`
	DestinationName string   `json:"destination_name"`
	Tags            []string `json:"tags"`
}

type WeChatTransaction struct {
	TransactionTime string `csv:"交易时间"`
	TransactionType string `csv:"交易类型"`
	Counterparty    string `csv:"交易对方"`
	Goods           string `csv:"商品"`
	InOrOut         string `csv:"收/支"`
	Amount          string `csv:"金额(元)"`
	PaymentMethod   string `csv:"支付方式"`
	Status          string `csv:"当前状态"`
	TransactionID   string `csv:"交易单号"`
	MerchantOrderID string `csv:"商户单号"`
	Remark          string `csv:"备注"`
}

type AlipayTransaction struct {
	TransactionTime     string `csv:"交易时间"`
	TransactionCategory string `csv:"交易分类"`
	Counterparty        string `csv:"交易对方"`
	CounterpartyAccount string `csv:"对方账号"`
	ProductDescription  string `csv:"商品说明"`
	InOrOut             string `csv:"收/支"`
	Amount              string `csv:"金额"`
	PaymentMethod       string `csv:"收/付款方式"`
	TransactionStatus   string `csv:"交易状态"`
	TransactionID       string `csv:"交易订单号"`
	MerchantOrderID     string `csv:"商家订单号"`
	Notes               string `csv:"备注"`
}

type ICBCTransaction struct {
	TransactionTime     string `csv:"交易日期"`
	AccountNumber       string `csv:"账号"`
	AccountType         string `csv:"储种"`
	SerialNumber        string `csv:"序号"`
	Currency            string `csv:"币种"`
	CashExchange        string `csv:"钞汇"`
	Summary             string `csv:"摘要"`
	Region              string `csv:"地区"`
	IncomeAmount        string `csv:"收入/支出金额"`
	Balance             string `csv:"余额"`
	Counterparty        string `csv:"对方户名"`
	CounterpartyAccount string `csv:"对方账号"`
	Channel             string `csv:"渠道"`
}

type TransactionError struct {
	WeChatTransactionError WeChatTransaction
	AlipayTransactionError AlipayTransaction
	ErrorInfo              string
}
