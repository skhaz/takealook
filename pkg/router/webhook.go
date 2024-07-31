package router

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/webhook"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
	"skhaz.dev/urlshortnen/pkg/mailer"
)

var secret = os.Getenv("STRIPE_WEBHOOK_SECRET")

func (r *Router) Webhook(c echo.Context) error {
	var (
		payload   []byte
		err       error
		signature string
		message   string
		event     stripe.Event
		invoice   stripe.Invoice
	)

	payload, err = io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error("error reading request body", zap.Error(err))
		return echo.NewHTTPError(http.StatusServiceUnavailable, "error reading request body")
	}

	signature = c.Request().Header.Get("stripe-signature")

	event, err = webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		// message = "error verifying webhook signature"
		// log.Error(message, zap.Error(err), zap.ByteString("payload", payload))
		return c.JSON(http.StatusOK, message)
	}

	switch event.Type {

	case "payment_intent.succeeded":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			message = "error parsing webhook json"
			log.Error(message, zap.Error(err))
			return c.JSON(http.StatusInternalServerError, message)
		}

		cm, err := customer.Get(invoice.Customer.ID, &stripe.CustomerParams{})
		if err != nil {
			message = "error getting user"
			log.Error(message, zap.Error(err))
			return c.JSON(http.StatusInternalServerError, message)
		}

		log.Info("payment_intent.succeeded for user", zap.String("email", cm.Email))

		if _, err = r.db.Exec(`UPDATE users SET active = 1 WHERE email = ?;`, cm.Email); err != nil {
			log.Fatal("failed to update user active status", zap.Error(err))
		}

		//nolint:golint,errcheck
		go mailer.NewMail().Send("rodrigo@delduca.org", "New subscription", cm.Email)

	case "payment_intent.canceled":
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			message = "error parsing webhook json"
			log.Error(message, zap.Error(err))
			return c.JSON(http.StatusInternalServerError, message)
		}

		cm, err := customer.Get(invoice.Customer.ID, &stripe.CustomerParams{})
		if err != nil {
			message = "error getting user"
			log.Error(message, zap.Error(err))
			return c.JSON(http.StatusInternalServerError, message)
		}

		log.Info("payment_intent.canceled for user", zap.String("email", cm.Email))

		if _, err = r.db.Exec(`UPDATE users SET active = 0 WHERE email = ?;`, cm.Email); err != nil {
			log.Fatal("failed to update user active status", zap.Error(err))
		}

		//nolint:golint,errcheck
		go mailer.NewMail().Send("rodrigo@delduca.org", "User canceled", cm.Email)
	}

	return c.NoContent(http.StatusOK)
}
