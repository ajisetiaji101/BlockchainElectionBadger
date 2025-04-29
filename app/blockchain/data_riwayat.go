package blockchain

type DataRiwayat struct {
	UniqueCodePayment string `json:"unique_code_payment"`
	Amount            int    `json:"amount"`
	Tax               string `json:"tax"`
	WarungCode        string `json:"warung_code"`
	Payment           string `json:"payment"`
	Date              string `json:"date"`
	Nik               string `json:"nik"`
	NikCustomer       string `json:"nik_customer"`
	Timestamp         int64  `json:"timestamp"`
}
