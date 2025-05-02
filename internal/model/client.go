package model

import (
	"time"
)

type Client struct {
	IP         string    // IP-адрес клиента
	Tokens     int       // текущее количество токенов
	Capacity   int       // максимальная емкость bucket
	Rate       int       // скорость пополнения токенов (токенов в единицу интервала)
	LastRefill time.Time // время последнего пополнения
}
