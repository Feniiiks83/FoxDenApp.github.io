package debtorControllerDto

type CreateDebtorBodyDto struct {
	Inn      string `json:"debtor_inn"`
	Ogrnip   string `json:"debtor_ogrnip"`
	Name     string `json:"debtor_name"`
	Category string `json:"debtor_category"`
	Snils    string `json:"debtor_snils"`
	Region   string `json:"debtor_region"`
	Address  string `json:"debtor_address"`
}

type UpdateDebtorData struct {
	Id       uint64 `json:"debtor_id" default:""`
	Inn      string `json:"debtor_inn" default:""`
	Ogrnip   string `json:"debtor_ogrnip" default:""`
	Name     string `json:"debtor_name" default:""`
	Category string `json:"debtor_category" default:""`
	Snils    string `json:"debtor_snils" default:""`
	Region   string `json:"debtor_region" default:""`
	Address  string `json:"debtor_address" default:""`
}
