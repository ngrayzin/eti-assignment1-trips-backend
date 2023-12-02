package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Response struct {
	Message string `json:"message"`
}

type Trips struct {
	TripID             int            `json:"tripID"`
	OwnerUserID        int            `json:"ownerUserID"`
	PickupLocation     string         `json:"pickupLoc"`
	AltPickupLocation  sql.NullString `json:"altPickupLoc"`
	StartTravelTime    string         `json:"startTravelTime"`
	DestinationAddress string         `json:"destinationAddress"`
	AvailableSeats     int            `json:"availableSeats"`
	IsActive           bool           `json:"isActive"`
	IsCancelled        bool           `json:"isCancelled"`
	IsStarted          bool           `json:"isStarted"`
	TripEndTime        sql.NullString `json:"tripEndTime"`
	CreatedAt          string         `json:"createdAt"`
	LastUpdated        sql.NullString `json:"lastUpdated"`
}

type TripEnrollment struct {
	EnrolmentID    int
	TripID         int
	PassengerID    int
	EnrollmentTime string
}

var db *sql.DB

var cfg = mysql.Config{
	User:      "user",
	Passwd:    "password",
	Net:       "tcp",
	Addr:      "localhost:3306",
	DBName:    "carpooling_db",
	ParseTime: true,
}

func main() {
	allowOrigins := handlers.AllowedOrigins([]string{"*"})
	allowMethod := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	db, _ = sql.Open("mysql", cfg.FormatDSN())
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/trips", trips).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/trips/{id}/{userid}", trips).Methods(http.MethodPut, http.MethodPost)
	router.HandleFunc("/api/v1/myEnrolments/{id}", myEnrolments).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/publishTrip", publishTrip).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/publishTrip/{id}", publishTrip).Methods(http.MethodPut)

	fmt.Println("Listening at port 5001")
	log.Fatal(http.ListenAndServe(":5001", handlers.CORS(allowHeaders, allowMethod, allowOrigins)(router)))
}

func trips(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println("/api/v1/trips")
		results, err := db.Query("SELECT * FROM Trips;")
		if err != nil {
			panic(err.Error())
		}
		defer results.Close()

		var trips []Trips
		for results.Next() {
			var trip Trips
			err = results.Scan(&trip.TripID, &trip.OwnerUserID, &trip.PickupLocation, &trip.AltPickupLocation, &trip.StartTravelTime, &trip.DestinationAddress, &trip.AvailableSeats, &trip.IsActive, &trip.IsCancelled, &trip.IsStarted, &trip.TripEndTime, &trip.CreatedAt, &trip.LastUpdated)
			if err != nil {
				panic(err.Error())
			}
			trips = append(trips, trip)
		}
		tripsJSON, err := json.Marshal(trips)
		if err != nil {
			panic(err.Error())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(tripsJSON)
	case http.MethodPut:
		params := mux.Vars(r)
		if _, ok := params["id"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No ID")
			return
		}
		if _, ok := params["userid"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No ID")
			return
		}
		id, _ := strconv.Atoi(params["id"])
		userid, _ := strconv.Atoi(params["userid"])

		var updateFields map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&updateFields); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid request body")
			return
		}

		fmt.Printf("/api/v1/trips/%d\n", id)

		result, err := db.Exec("INSERT INTO TripEnrollments (TripID, PassengerUserID) VALUES (?, ?)", id, userid)
		if err != nil {
			me, ok := err.(*mysql.MySQLError)
			if !ok {
				panic(err.Error())
			}
			if me.Number == 1062 {
				fmt.Println("Already have this enrolment")
				fmt.Fprintf(w, "Duplicate\n")
				w.WriteHeader(http.StatusConflict)
				return
			}
		}

		lastInsertID, err := result.LastInsertId()
		if err != nil {
			panic(err.Error())
		}

		var setClauses []string
		var values []interface{}

		for key, value := range updateFields {
			if key == "IsActive" || key == "IsCancelled" || key == "IsStarted" {
				setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
				// Check if the value is a string representation of a boolean
				if strValue, ok := value.(string); ok {
					// Convert string representation to boolean
					boolValue, err := strconv.ParseBool(strValue)
					if err != nil {
						boolValue = false // Default value set to false
					}
					values = append(values, boolValue)
				} else if boolValue, ok := value.(bool); ok {
					values = append(values, boolValue)
				} else {
					values = append(values, value)
				}
			} else {
				setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
				values = append(values, value)
			}
		}

		query := fmt.Sprintf(`
			UPDATE Trips
			SET %s
			WHERE TripID = ?;
		`, strings.Join(setClauses, ", "))

		values = append(values, id)
		rows, err := db.Query(query, values...)
		if err != nil {
			panic(err.Error())
		}
		defer rows.Close()

		fmt.Printf("Trip with id %d updated\n", id)
		fmt.Fprintf(w, "Trip data updated successfully\n")
		fmt.Printf("Enrollment with id %d completed for user with id %d\n", lastInsertID, userid)
		fmt.Fprintf(w, "Enrollment success\n")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Error")
	}
}

func myEnrolments(w http.ResponseWriter, r *http.Request) {
	type TripEnrollmentData struct {
		Email              string         `json:"email"`
		FirstName          string         `json:"firstName"`
		LastName           string         `json:"lastName"`
		MobileNumber       string         `json:"mobileNumber"`
		PickupLocation     string         `json:"pickupLocation"`
		AltPickupLocation  sql.NullString `json:"altPickupLocation"`
		StartTravelTime    string         `json:"startTravelTime"`
		DestinationAddress string         `json:"destinationAddress"`
		IsActive           bool           `json:"isActive"`
		IsCancelled        bool           `json:"isCancelled"`
		IsStarted          bool           `json:"isStarted"`
		TripEndTime        sql.NullString `json:"tripEndTime"`
		CreatedAt          string         `json:"createdAt"`
	}

	params := mux.Vars(r)
	if _, ok := params["id"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No ID")
		return
	}

	id, _ := strconv.Atoi(params["id"])
	fmt.Printf("/api/v1/myEnrolments/%d\n", id)
	results, err := db.Query("SELECT u.Email, u.FirstName, u.Lastname, u.MobileNumber, t.PickupLocation,t.AltPickupLocation, t.StartTravelTime, t.DestinationAddress, t.IsActive, t.IsCancelled, t.IsStarted, t.TripEndTime, t.CreatedAt FROM ((Trips t INNER JOIN TripEnrollments te ON t.TripID = te.TripID) INNER JOIN Users u ON t.OwnerUserID = u.UserID) WHERE PassengerUserID = ?;", id)
	if err != nil {
		panic(err.Error())
	}
	defer results.Close()
	var enrollments []TripEnrollmentData
	for results.Next() {
		var e TripEnrollmentData
		err = results.Scan(&e.Email, &e.FirstName, &e.LastName, &e.MobileNumber, &e.PickupLocation, &e.AltPickupLocation, &e.StartTravelTime, &e.DestinationAddress, &e.CreatedAt)
		if err != nil {
			panic(err.Error())
		}
		enrollments = append(enrollments, e)
	}
	enrollmentsJSON, err := json.Marshal(enrollments)
	if err != nil {
		panic(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(enrollmentsJSON)
}

func publishTrip(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var trip Trips
		err := decoder.Decode(&trip)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if trip.OwnerUserID <= 0 || trip.PickupLocation == "" || trip.StartTravelTime == "" || trip.DestinationAddress == "" || trip.AvailableSeats <= 0 {
			fmt.Println("Invalid params")
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
			return
		}

		fmt.Println("/api/v1/publishTrip")

		result, err := db.Exec("INSERT INTO Trips (OwnerUserID, PickupLocation, StartTravelTime, DestinationAddress, AvailableSeats) VALUES (?, ?, ?, ?, ?)",
			trip.OwnerUserID, trip.PickupLocation, trip.StartTravelTime, trip.DestinationAddress, trip.AvailableSeats)
		if err != nil {
			panic(err.Error())
		}

		id, err := result.LastInsertId()
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("Trip created with ID:", id)
		fmt.Fprintf(w, "Trip created with ID: %d", id)
	case http.MethodPut:
		params := mux.Vars(r)
		if _, ok := params["id"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No ID")
			return
		}
		id, _ := strconv.Atoi(params["id"])
		var updateFields map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&updateFields); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid request body")
			return
		}

		fmt.Printf("/api/v1/publishTrip/%d", id)

		var setClauses []string
		var values []interface{}

		for key, value := range updateFields {
			if key == "IsActive" || key == "IsCancelled" || key == "IsStarted" {
				setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
				// Check if the value is a string representation of a boolean
				if strValue, ok := value.(string); ok {
					// Convert string representation to boolean
					boolValue, err := strconv.ParseBool(strValue)
					if err != nil {
						boolValue = false // Default value set to false
					}
					values = append(values, boolValue)
				} else if boolValue, ok := value.(bool); ok {
					values = append(values, boolValue)
				} else {
					values = append(values, value)
				}
			} else {
				setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
				values = append(values, value)
			}
		}

		query := fmt.Sprintf(`
			UPDATE Trips
			SET %s
			WHERE TripID = ?;
		`, strings.Join(setClauses, ", "))

		values = append(values, id)
		rows, err := db.Query(query, values...)
		if err != nil {
			panic(err.Error())
		}
		defer rows.Close()

		fmt.Printf("Trip with id %d updated\n", id)
		fmt.Fprintf(w, "Trip data updated successfully\n")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Error")
	}
}