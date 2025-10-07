package delete

import (
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

// конструктор для handler удаления URL
func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// получаем alias из URL параметров
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		log.Info("attempting to delete url", slog.String("alias", alias))

		// удаляем URL по alias
		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if err == storage.ErrURLNotFound {
				log.Info("url not found", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("url not found"))
				return
			}

			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to delete url"))
			return
		}

		log.Info("url deleted", slog.String("alias", alias))
		render.JSON(w, r, resp.OK())
	}
}
