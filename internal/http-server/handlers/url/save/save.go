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
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		// ДОБАВКА: отдельная обработка коллизии алиаса
		if errors.Is(err, storage.ErrURLExists) {
			// если алиас задан пользователем — сразу конфликт
			if req.Alias != "" {
				render.JSON(w, r, resp.Error("alias already exists"))
				return
			}

			// если алиас сгенерирован — пробуем несколько раз
			for attempt := 1; attempt <= 4; attempt++ {
				alias = random.NewRandomString(aliasLenght)
				if _, err = urlSaver.SaveURL(req.URL, alias); err == nil {
					render.JSON(w, r, Response{Response: resp.OK(), Alias: alias})
					return
				}
				if !errors.Is(err, storage.ErrURLExists) {
					render.JSON(w, r, resp.Error("failed to save url"))
					return
				}
			}
			render.JSON(w, r, resp.Error("could not generate unique alias"))
			return
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		// прочие неожиданные ошибки
		render.JSON(w, r, resp.Error("failed to save url"))

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
