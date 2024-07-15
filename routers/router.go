package routers

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"test-project/controllers"
)

func InitRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.HandleFunc("/users", controllers.GetUsers).Methods("GET")

	router.HandleFunc("/users", controllers.CreateUser).Methods("POST")

	router.HandleFunc("/users/{id}", controllers.UpdateUser).Methods("PATCH")

	router.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE")

	router.HandleFunc("/users/{id}/tasks", controllers.GetUserTasks).Methods("GET")

	router.HandleFunc("/users/{id}/tasks/start", controllers.StartTask).Methods("POST")

	router.HandleFunc("/users/{id}/tasks/stop", controllers.StopTask).Methods("POST")

	return router
}
