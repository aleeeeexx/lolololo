package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestToPublicGroupRateDTOs_LocalizesNameAndKeepsRates(t *testing.T) {
	groups := []service.PublicGroupRate{
		{
			ID:             2,
			Name:           "default",
			I18n:           service.GroupI18n{"zh": {Name: "默认分组"}},
			RateMultiplier: 1,
		},
		{
			ID:                   3,
			Name:                 "image-heavy",
			RateMultiplier:       0.5,
			ImageRateIndependent: true,
			ImageRateMultiplier:  1.2,
			VideoRateIndependent: true,
			VideoRateMultiplier:  0.2,
		},
	}

	zh := toPublicGroupRateDTOs(groups, "zh-CN")
	require.Equal(t, "默认分组", zh[0].Name)
	require.Equal(t, 1.0, zh[0].RateMultiplier)
	require.Equal(t, "image-heavy", zh[1].Name)
	require.True(t, zh[1].ImageRateIndependent)
	require.Equal(t, 1.2, zh[1].ImageRateMultiplier)
	require.True(t, zh[1].VideoRateIndependent)
	require.Equal(t, 0.2, zh[1].VideoRateMultiplier)

	en := toPublicGroupRateDTOs(groups, "en")
	require.Equal(t, "default", en[0].Name)
}

func TestGetPublicGroupRates_NilServiceIsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/group-rates", nil)

	(&ModelPlazaHandler{}).GetPublicGroupRates(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	var payload response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.NotEqual(t, 0, payload.Code)
}

func TestToPublicGroupRateDTOs_EmptyInput(t *testing.T) {
	out := toPublicGroupRateDTOs(nil, "zh")
	require.NotNil(t, out)
	require.Empty(t, out)
}
