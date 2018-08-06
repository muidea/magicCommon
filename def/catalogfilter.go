package def

import (
	"fmt"
	"net/http"
	"strconv"

	"muidea.com/magicCommon/model"
)

// DecodeCatalog 解析CatalogUnit
func DecodeCatalog(request *http.Request) (*model.CatalogUnit, error) {
	ret := &model.CatalogUnit{}
	idStr := request.URL.Query().Get("strictID")
	typeStr := request.URL.Query().Get("strictType")
	if len(idStr) > 0 || len(typeStr) > 0 {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, err
		}

		ret.ID = id
		ret.Type = typeStr
		return ret, nil
	}

	return nil, nil
}

// EncodeCatalog 对catalog进行编码
func EncodeCatalog(catalog model.CatalogUnit) string {
	return fmt.Sprintf("strictID=%d&strictType=%s", catalog.ID, catalog.Type)
}
