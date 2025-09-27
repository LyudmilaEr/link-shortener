package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// TODO: move to config (or to DB)
const aliasLenght = 6

// go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave, alias string) (int64, error)
}

// конструктор для handler, будет вызываться при подклчении к роутеру
func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// валидация
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			// render.JSON(w, r, resp.Error("invalid request"))
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLenght)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if err == nil {
			log.Info("url added", slog.Int64("id", id))
			responseOK(w, r, alias)
			return
		}

		// Handle alias collision
		if errors.Is(err, storage.ErrURLExists) {
			// если алиас задан пользователем — сразу конфликт
			if req.Alias != "" {
				log.Info("alias already in use", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("alias already exists"))
				return
			}

			// если алиас сгенерирован — пробуем несколько раз
			for attempt := 1; attempt <= 4; attempt++ {
				alias = random.NewRandomString(aliasLenght)
				if id, err = urlSaver.SaveURL(req.URL, alias); err == nil {
					log.Info("url saved after retry", slog.Int64("id", id), slog.String("alias", alias), slog.Int("attempt", attempt))
					responseOK(w, r, alias)
					return
				}
				if !errors.Is(err, storage.ErrURLExists) {
					log.Error("failed to save url", sl.Err(err))
					render.JSON(w, r, resp.Error("failed to save url"))
					return
				}
				log.Info("generated alias collision, retrying", slog.String("alias", alias), slog.Int("attempt", attempt))
			}
			// Exhausted attempts to generate a unique alias
			log.Error("could not generate unique alias")
			render.JSON(w, r, resp.Error("could not generate unique alias"))
			return
		}

		// Unexpected storage error
		log.Error("failed to save url", sl.Err(err))
		render.JSON(w, r, resp.Error("failed to save url"))
		return
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
