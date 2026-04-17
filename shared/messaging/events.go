package messaging

import (
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue        = "q.find_available_drivers"
	DriverCmdTripRequestQueue        = "q.driver_cmd_trip"
	DriverTripResponseQueue          = "q.driver_trip_response"
	NotifyDriverNotFoundQueue        = "q.notify_driver_not_found"
	NotifyDriverAssignQueue          = "q.notify_driver_assign"
	NotifyPaymentSessionCreatedQueue = "q.notify_payment_session_created"
	PaymentTripResponseQueue         = "q.payment_trip_response"
	NotifyPaymentSuccessQueue        = "q.notify_payment_success"

	DeadLetterQueue = "q.dead_letter"
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

type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripID"`
	SessionID string  `json:"sessionID"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentTripResponseData struct {
	TripID   string  `json:"tripID"`
	UserID   string  `json:"userID"`
	DriverID string  `json:"driverID"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentStatusUpdateData struct {
	TripID   string `json:"tripID"`
	UserID   string `json:"userID"`
	DriverID string `json:"driverID"`
}
