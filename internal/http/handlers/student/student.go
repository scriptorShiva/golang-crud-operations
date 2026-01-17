package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/scriptorshiva/golang-crud-operations/internal/storage"

	"github.com/scriptorshiva/golang-crud-operations/internal/types"
	"github.com/scriptorshiva/golang-crud-operations/internal/utils/response"
)

// create method - In Go convention : New
// here in New() - we are using DI
func New(storage storage.Storage) http.HandlerFunc {
	// we will return handler function
	return func(rw http.ResponseWriter, r *http.Request){

		slog.Info("creating student")

		// In go - we have to decode to get information from req body , we have to serialize to get information from req body in struct
		var student types.Student

		// any json data , decode it to struct student. It return err
		err := json.NewDecoder(r.Body).Decode(&student)
		// If body empty error
		if errors.Is(err , io.EOF){
			// sending 400 
			response.WriteJson(rw, http.StatusBadRequest, response.CommonError(fmt.Errorf("body is empty")))
			return
		}

		if err != nil {
			// sending 400 
			response.WriteJson(rw, http.StatusBadRequest, response.CommonError(fmt.Errorf("unable to decode request body: %w", err)))
			return
		}

		// validate request (0 trust policy on client)
		if err := validator.New().Struct(student); err != nil {
			// typecast error
			validateErr := err.(validator.ValidationErrors)
			response.WriteJson(rw, http.StatusBadRequest, response.ValidationError(validateErr))
			return
		}

		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		slog.Info("student created", slog.Int64("id", lastId))

		if err != nil {
			response.WriteJson(rw, http.StatusInternalServerError, response.CommonError(err))
			return
		}

		response.WriteJson(rw , http.StatusCreated, map[string]int64 {"id": lastId})

	}
}

func FetchById(storage storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		slog.Info("fetching student by id")

		id := r.PathValue("id")
		// parse string to int
		idInt, err := strconv.Atoi(id)

		if err != nil {
			response.WriteJson(rw, http.StatusBadRequest, response.CommonError(err))
			return
		}

		student, err := storage.FetchStudentById(idInt)

		if err != nil {
			response.WriteJson(rw, http.StatusInternalServerError, response.CommonError(err))
			return
		}

		response.WriteJson(rw,http.StatusOK, student)

	}
}

func FetchAll(storage storage.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		slog.Info("fetching all students")

		students, err := storage.FetchAllStudents()

		if err != nil {
			response.WriteJson(rw, http.StatusInternalServerError, response.CommonError(err))
			return
		}

		response.WriteJson(rw,http.StatusOK, students)
	}
}