package client

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/gofrs/uuid"

	"github.com/andranikuz/gophkeeper/pkg/entity"
)

// CardDTO представляет подробные данные карты.
type CardDTO struct {
	CardNumber     string `json:"card_number"`
	ExpirationDate string `json:"expiration_date"` // Формат MM/YY или MM/YYYY
	CVV            string `json:"cvv"`
	CardHolderName string `json:"card_holder_name"`
	Meta           string `json:"meta"`
}

// Validate выполняет валидацию данных карты.
func (dto *CardDTO) Validate() error {
	// Проверка номера карты: от 13 до 19 цифр.
	numRe := regexp.MustCompile(`^\d{13,19}$`)
	if !numRe.MatchString(dto.CardNumber) {
		return errors.New("invalid card number: must be 13 to 19 digits")
	}

	// Проверка формата срока действия.
	var exp time.Time
	var err error
	if len(dto.ExpirationDate) == 5 {
		// Формат MM/YY
		exp, err = time.Parse("01/06", dto.ExpirationDate)
	} else if len(dto.ExpirationDate) == 7 {
		// Формат MM/YYYY
		exp, err = time.Parse("01/2006", dto.ExpirationDate)
	} else {
		return errors.New("expiration date must be in format MM/YY or MM/YYYY")
	}
	if err != nil {
		return errors.New("invalid expiration date format")
	}
	// Допустим карта действительна до конца месяца.
	year, month, _ := exp.Date()
	expLastDay := time.Date(year, month+1, 0, 23, 59, 59, 0, time.Local)
	if time.Now().After(expLastDay) {
		return errors.New("card is expired")
	}

	// Проверка CVV (3 или 4 цифры).
	cvvRe := regexp.MustCompile(`^\d{3,4}$`)
	if !cvvRe.MatchString(dto.CVV) {
		return errors.New("invalid CVV: must be 3 or 4 digits")
	}

	// Проверка наличия имени владельца.
	if dto.CardHolderName == "" {
		return errors.New("card holder name cannot be empty")
	}
	return nil
}

// SaveCard сохраняет данные типа "card" в локальное хранилище.
// Сначала выполняется валидация DTO, затем сериализация, кодирование в base64 и сохранение.
func (c *Client) SaveCard(ctx context.Context, dto CardDTO) error {
	if err := dto.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	id, err := uuid.NewV6()
	if err != nil {
		return err
	}
	return c.LocalDB.SaveItem(entity.NewDataItem(
		id.String(),
		entity.DataTypeCard,
		string(payload),
		dto.Meta,
		c.Session.GetUserID(),
	))
}
