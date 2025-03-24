package delivery

import (
	deliveryHTTP "go-crud/delivery/http"
	"go-crud/internal/usecase"
	"github.com/go-chi/chi/v5"
	"net/http"
)
func NewRouter(userUC usecase.IUserUsecase, repoUC usecase.IRepositoryUsecase) *chi.Mux {
	r := chi.NewRouter()

	// User handler
	userHandler := deliveryHTTP.NewUserHandler(userUC)
	r.Post("/users", userHandler.CreateUser)
	r.Get("/users", userHandler.GetAllUsers)
	r.Get("/users/{id}", userHandler.GetUserByID)
	r.Put("/users/{id}", userHandler.UpdateUser)  
	r.Delete("/users/{id}", userHandler.DeleteUser) 

	// Repository handler
	repoHandler := deliveryHTTP.NewRepositoryHandler(repoUC)
	r.Post("/users/{id}/repositories", repoHandler.CreateRepository)
	r.Get("/users/{id}/repositories", repoHandler.GetRepositoriesByUserID) 
	r.Get("/repositories/{id}", repoHandler.GetRepositoryByID) 
	r.Put("/repositories/{id}", repoHandler.UpdateRepository) 
	r.Delete("/repositories/{id}", repoHandler.DeleteRepository) 
	// Health Check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is running..."))
	})

	return r
}
