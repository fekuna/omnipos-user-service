package merchant

import (
	"github.com/fekuna/omnipos-user-service/internal/merchant/dto"
	"github.com/fekuna/omnipos-user-service/internal/model"
)

type PGRepository interface {
	FindOneByAttributes(input *dto.FindOneByAttribute) (*model.Merchant, error)
}
