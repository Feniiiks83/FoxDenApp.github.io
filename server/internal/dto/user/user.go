package user_dto

import (
	"time"
)

type CreateUserResponse struct {
	Time time.Time `json:"time`
	Ip   string    `json:"ip"`
}

// type FindAllDebtorsReturning struct {
// 	Count   int            `json:"count"`
// 	Debtors *[]models.User `json:"debtors"`
// }
