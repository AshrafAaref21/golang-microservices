package messaging

import (
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "q.find_available_drivers"
	DriverCmdTripRequestQueue = "q.driver_cmd_trip"
	DriverTripResponseQueue   = "q.driver_trip_response"
	NotifyDriverNotFoundQueue = "q.notify_driver_not_found"
	NotifyDriverAssignQueue   = "q.notify_driver_assign"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
	// Accepted bool       `json:"accepted"`
}
