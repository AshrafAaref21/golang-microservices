package messaging

import pb "ride-sharing/shared/proto/trip"

const (
	FindAvailableDriversQueue = "q.find_available_drivers"
	DriverCmdTripRequestQueue = "q.driver_cmd_trip"
	DriverTripResponseQueue   = "q.driver_trip_response"
	NotifyDriverNotFoundQueue = "q.notify_driver_not_found"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
