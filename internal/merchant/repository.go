package merchant

import (
	"context"

	"github.com/fekuna/omnipos-user-service/internal/merchant/dto"
	"github.com/fekuna/omnipos-user-service/internal/model"
)

type PGRepository interface {
	FindOneByAttributes(ctx context.Context, input *dto.FindOneByAttribute) (*model.Merchant, error)
}
