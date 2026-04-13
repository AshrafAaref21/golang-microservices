package messaging

import pb "ride-sharing/shared/proto/trip"

const (
	FindAvailableDriversQueue = "q.find_available_drivers"
	DriverCmdTripRequestQueue = "q.driver_cmd_trip"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
