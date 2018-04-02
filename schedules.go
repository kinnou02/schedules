package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/kinnou02/gonavitia"
	"github.com/kinnou02/gonavitia/pbnavitia"
	"github.com/kinnou02/gonavitia/serializer"
	log "github.com/sirupsen/logrus"
)

type RouteScheduleRequest struct {
	FromDatetime     time.Time     `form:"from_datetime" time_format:"20060102150405" binding:"required"`
	DisableGeojson   bool          `form:"disable_geojson"`
	StartPage        int32         `form:"start_page"`
	Count            int32         `form:"count"`
	Duration         time.Duration `form:"duration"`
	ForbiddenUris    []string      //mapping with Binding doesn't work
	Depth            int32         `form:"depth"`
	CurrentDatetime  time.Time     `form:"_current_datetime"`
	ItemsPerSchedule int32
	Filters          []string
}

func NewRouteScheduleRequest() RouteScheduleRequest {
	return RouteScheduleRequest{
		StartPage:        0,
		Count:            10,
		Duration:         24 * time.Hour,
		CurrentDatetime:  time.Now(),
		Depth:            2,
		ItemsPerSchedule: 10000,
	}
}

func RouteSchedule(c *gin.Context, kraken *gonavitia.Kraken, request *RouteScheduleRequest) {
	pb_req := BuildRequestRouteSchedule(*request)
	resp, err := kraken.Call(pb_req)
	if err != nil {
		log.Errorf("FATAL: %+v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	r := serializer.NewRouteSchedulesResponse(resp)
	c.JSON(http.StatusOK, r)
}

func BuildRequestRouteSchedule(req RouteScheduleRequest) *pbnavitia.Request {
	departureFilter := strings.Join(req.Filters, "and ")
	pb_req := &pbnavitia.Request{
		RequestedApi: pbnavitia.API_ROUTE_SCHEDULES.Enum(),
		NextStopTimes: &pbnavitia.NextStopTimeRequest{
			DepartureFilter:  proto.String(departureFilter),
			ArrivalFilter:    proto.String(""),
			FromDatetime:     proto.Uint64(uint64(req.FromDatetime.Unix())),
			Duration:         proto.Int32(int32(req.Duration.Seconds())),
			Depth:            proto.Int32(req.Depth),
			NbStoptimes:      proto.Int32(req.Count),
			Count:            proto.Int32(req.Count),
			StartPage:        proto.Int32(req.StartPage),
			DisableGeojson:   proto.Bool(req.DisableGeojson),
			ItemsPerSchedule: proto.Int32(req.ItemsPerSchedule),
			RealtimeLevel:    pbnavitia.RTLevel_BASE_SCHEDULE.Enum(),
		},
		XCurrentDatetime: proto.Uint64(uint64(req.CurrentDatetime.Unix())),
	}
	pb_req.NextStopTimes.ForbiddenUri = append(pb_req.NextStopTimes.ForbiddenUri, req.ForbiddenUris...)

	return pb_req
}
